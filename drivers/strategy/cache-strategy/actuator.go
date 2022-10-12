package cache_strategy

import (
	"github.com/eolinker/apinto/drivers/strategy/cache-strategy/cache"
	"github.com/eolinker/apinto/strategy"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
)

var (
	actuatorSet ActuatorSet
)

func init() {
	actuator := newtActuator()
	actuatorSet = actuator
	strategy.AddStrategyHandler(actuator)
	cache.NewCache()
}

type ActuatorSet interface {
	Set(string, *CacheValidTimeHandler)
	Del(id string)
}

type tActuator struct {
	lock     sync.RWMutex
	all      map[string]*CacheValidTimeHandler
	handlers []*CacheValidTimeHandler
}

func (a *tActuator) Destroy() {

}

func (a *tActuator) Set(id string, val *CacheValidTimeHandler) {
	// 调用来源有锁
	a.all[id] = val
	a.rebuild()

}

func (a *tActuator) Del(id string) {
	// 调用来源有锁
	delete(a.all, id)
	a.rebuild()
}

func (a *tActuator) rebuild() {

	handlers := make([]*CacheValidTimeHandler, 0, len(a.all))
	for _, h := range a.all {
		if !h.stop {
			handlers = append(handlers, h)
		}
	}
	sort.Sort(handlerListSort(handlers))
	a.lock.Lock()
	defer a.lock.Unlock()
	a.handlers = handlers
}
func newtActuator() *tActuator {
	return &tActuator{
		all: make(map[string]*CacheValidTimeHandler),
	}
}

func (a *tActuator) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) error {

	httpCtx, err := http_service.Assert(ctx)
	if err != nil {
		return err
	}

	if httpCtx.Request().Method() != http.MethodGet {
		return next.DoChain(ctx)
	}
	a.lock.RLock()
	handlers := a.handlers
	a.lock.RUnlock()

	for _, handler := range handlers {
		if handler.stop {
			continue
		}
		if handler.filter.Check(httpCtx) {

			uri := httpCtx.Request().URI().RequestURI()
			responseData := cache.GetResponseData(uri)

			if responseData != nil {
				httpCtx.SetCompleteHandler(responseData)
			} else {
				httpCtx.SetCompleteHandler(NewCacheGetCompleteHandler(httpCtx.GetComplete(), handler.validTime))
			}
			break
		}
	}

	if next != nil {
		return next.DoChain(ctx)
	}
	return nil
}

type CacheGetCompleteHandler struct {
	orgHandler eocontext.CompleteHandler
	validTime  int
}

func NewCacheGetCompleteHandler(orgHandler eocontext.CompleteHandler, validTime int) *CacheGetCompleteHandler {
	return &CacheGetCompleteHandler{orgHandler: orgHandler, validTime: validTime}
}

func (c *CacheGetCompleteHandler) Complete(ctx eocontext.EoContext) error {

	err := c.orgHandler.Complete(ctx)
	if err != nil {
		return err
	}
	httpCtx, err2 := http_service.Assert(ctx)
	if err2 != nil {
		return nil
	}

	//从cache-control中判断是否需要缓存
	if parseCacheControl(httpCtx).IsCache() {
		header := make(map[string]string)
		for key, values := range httpCtx.Response().Headers() {
			if len(values) > 0 {
				header[key] = values[0]
			}
		}

		responseData := &cache.ResponseData{
			Header: header,
			Body:   httpCtx.Response().GetBody(),
		}
		cache.SetResponseData(httpCtx.Request().URI().RequestURI(), responseData, c.validTime)
	}
	return nil
}

type handlerListSort []*CacheValidTimeHandler

func (hs handlerListSort) Len() int {
	return len(hs)
}

func (hs handlerListSort) Less(i, j int) bool {

	return hs[i].priority < hs[j].priority
}

func (hs handlerListSort) Swap(i, j int) {
	hs[i], hs[j] = hs[j], hs[i]
}

type cacheControlMap map[string]string

func (c cacheControlMap) NoCache() bool {
	if _, ok := c["no-cache"]; ok {
		return true
	}
	return false
}

func (c cacheControlMap) MaxAge() int {
	if maxAgeStr, ok := c["max-age"]; ok {
		maxAge, _ := strconv.Atoi(maxAgeStr)
		return maxAge
	}
	return 0
}

func (c cacheControlMap) IsCache() bool {
	if c.MaxAge() == 0 {
		return false
	}
	if c.NoCache() {
		return false
	}
	if !c.IsPublic() {
		return false
	}
	if _, ok := c["no-store"]; ok {
		return false
	}
	return true
}

func (c cacheControlMap) IsPublic() bool {
	if _, ok := c["Authorization"]; ok {
		if _, pOk := c["public"]; pOk {
			return true
		} else {
			return false
		}
	}
	//只要不是私有的 都算公有
	if _, ok := c["private"]; ok {
		return false
	}
	return true
}

func parseCacheControl(httpCtx http_service.IHttpContext) cacheControlMap {
	cc := cacheControlMap{}

	header := httpCtx.Response().GetHeader("Cache-Control")
	for _, part := range strings.Split(header, ",") {
		part = strings.Trim(part, " ")
		if part == "" {
			continue
		}
		if strings.ContainsRune(part, '=') {
			keyVal := strings.Split(part, "=")
			cc[strings.Trim(keyVal[0], " ")] = strings.Trim(keyVal[1], ",")
		} else {
			cc[part] = ""
		}
	}
	return cc
}
