package fuse_strategy

import (
	"fmt"
	"github.com/eolinker/apinto/resources"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"sort"
	"sync"
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
	Set(string, *FuseHandler)
	Del(id string)
}

type tActuator struct {
	lock     sync.RWMutex
	all      map[string]*FuseHandler
	handlers []*FuseHandler
}

func (a *tActuator) Destroy() {

}

func (a *tActuator) Set(id string, val *FuseHandler) {
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

	handlers := make([]*FuseHandler, 0, len(a.all))
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
		all: make(map[string]*FuseHandler),
	}
}

func (a *tActuator) Strategy(ctx eocontext.EoContext, next eocontext.IChain, cache resources.ICache) error {

	httpCtx, err := http_service.Assert(ctx)
	if err != nil {
		return err
	}
	a.lock.RLock()
	handlers := a.handlers
	a.lock.RUnlock()

	for _, handler := range handlers {
		//check筛选条件
		if !handler.filter.Check(httpCtx) {
			continue
		}
		if handler.Fusing(ctx, cache) {
			res := handler.rule.response
			httpCtx.Response().SetStatus(res.statusCode, "")
			for _, h := range res.headers {
				httpCtx.Response().SetHeader(h.key, h.value)
			}
			httpCtx.Response().SetHeader("Content-Type", fmt.Sprintf("%s; charset=%s", res.contentType, res.charset))
			httpCtx.Response().SetBody([]byte(res.body))
			return nil
		}
	}

	if next != nil {
		return next.DoChain(ctx)
	}
	return nil
}

func (a *tActuator) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) error {

	httpCtx, err := http_service.Assert(ctx)
	if err != nil {
		return err
	}

	a.lock.RLock()
	handlers := a.handlers
	a.lock.RUnlock()

	for _, handler := range handlers {
		//check筛选条件
		if handler.filter.Check(httpCtx) {

			break
		}
	}

	if next != nil {
		return next.DoChain(ctx)
	}
	return nil
}

type handlerListSort []*FuseHandler

func (hs handlerListSort) Len() int {
	return len(hs)
}

func (hs handlerListSort) Less(i, j int) bool {

	return hs[i].priority < hs[j].priority
}

func (hs handlerListSort) Swap(i, j int) {
	hs[i], hs[j] = hs[j], hs[i]
}

func DoStrategy(ctx eocontext.EoContext, next eocontext.IChain, iCache resources.ICache) error {
	return actuatorSet.Strategy(ctx, next, iCache)
}
