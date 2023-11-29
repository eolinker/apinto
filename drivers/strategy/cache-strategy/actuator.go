package cache_strategy

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"

	"github.com/eolinker/apinto/drivers/strategy/cache-strategy/cache"
	"github.com/eolinker/apinto/resources"
)

var (
	actuatorSet ActuatorSet
)

func init() {
	actuator := newtActuator()
	actuatorSet = actuator
}

type ActuatorSet interface {
	Strategy(ctx eocontext.EoContext, next eocontext.IChain, cache resources.ICache) error
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

func (a *tActuator) Strategy(ctx eocontext.EoContext, next eocontext.IChain, iCache resources.ICache) error {

	httpCtx, err := http_service.Assert(ctx)
	if err != nil {
		if next != nil {
			return next.DoChain(ctx)
		}
		return err
	}

	if httpCtx.Request().Method() != http.MethodGet {
		if next != nil {
			return next.DoChain(ctx)
		}
		return nil
	}

	a.lock.RLock()
	handlers := a.handlers
	a.lock.RUnlock()

	for _, handler := range handlers {
		if handler.filter.Check(httpCtx) {

			uri := httpCtx.Request().URI().RequestURI()
			responseData := cache.GetResponseData(iCache, uri)

			if responseData != nil {
				ctx.SetLabel("handler", "cache")
				httpCtx.SetCompleteHandler(responseData)
			} else {
				httpCtx.SetCompleteHandler(NewCacheGetCompleteHandler(httpCtx.GetComplete(), handler.validTime, uri, iCache))
			}
			break
		}
	}

	if next != nil {
		return next.DoChain(ctx)
	}
	return nil
}

type CacheCompleteHandler struct {
	orgHandler eocontext.CompleteHandler
	validTime  int
	uri        string
	cache      resources.ICache
}

func NewCacheGetCompleteHandler(orgHandler eocontext.CompleteHandler, validTime int, uri string, cache resources.ICache) *CacheCompleteHandler {
	return &CacheCompleteHandler{
		orgHandler: orgHandler,
		validTime:  validTime,
		uri:        uri,
		cache:      cache,
	}
}

func (c *CacheCompleteHandler) Complete(ctx eocontext.EoContext) error {

	if c.orgHandler != nil {
		if err := c.orgHandler.Complete(ctx); err != nil {
			return err
		}
	}

	httpCtx, err := http_service.Assert(ctx)
	if err != nil {
		return nil
	}

	//从cache-control中判断是否需要缓存
	if parseHttpContext(httpCtx).IsCache() {
		responseData := &cache.ResponseData{
			Header:    httpCtx.Response().Headers(),
			Body:      httpCtx.Response().GetBody(),
			ValidTime: c.validTime,
			Now:       time.Now(),
		}
		cache.SetResponseData(c.cache, httpCtx.Request().URI().RequestURI(), responseData, c.validTime)
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

type httpContext struct {
	cacheControl map[string]string
	reqHeader    http.Header
	resHeader    http.Header
}

func (c httpContext) NoCache() bool {
	if _, ok := c.cacheControl["no-cache"]; ok {
		return true
	}
	return false
}

func (c httpContext) IsCache() bool {
	if maxAgeStr, ok := c.cacheControl["max-age"]; ok {
		maxAge, _ := strconv.Atoi(maxAgeStr)
		if maxAge == 0 {
			return false
		}
	}

	if c.NoCache() {
		return false
	}

	if !c.IsPublic() {
		return false
	}

	if _, ok := c.cacheControl["no-store"]; ok {
		return false
	}

	return true
}

func (c httpContext) IsPublic() bool {
	if _, ok := c.reqHeader["Authorization"]; ok {
		if _, pOk := c.cacheControl["public"]; pOk {
			return true
		}
		return false
	}

	//只要不是私有的 都算公有
	if _, ok := c.cacheControl["private"]; ok {
		return false
	}
	return true
}

func parseHttpContext(httpCtx http_service.IHttpContext) httpContext {
	hc := httpContext{
		cacheControl: make(map[string]string),
	}

	hc.resHeader = httpCtx.Response().Headers()
	hc.reqHeader = httpCtx.Request().Header().Headers()

	cacheControlHeader := httpCtx.Response().GetHeader("Cache-Control")
	for _, part := range strings.Split(cacheControlHeader, ",") {
		part = strings.Trim(part, " ")
		if part == "" {
			continue
		}
		if strings.ContainsRune(part, '=') {
			keyVal := strings.Split(part, "=")
			hc.cacheControl[strings.Trim(keyVal[0], " ")] = strings.Trim(keyVal[1], ",")

		} else {
			hc.cacheControl[part] = ""
		}
	}
	return hc
}

func DoStrategy(ctx eocontext.EoContext, next eocontext.IChain, cache resources.ICache) error {
	return actuatorSet.Strategy(ctx, next, cache)
}
