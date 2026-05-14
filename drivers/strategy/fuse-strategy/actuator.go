package fuse_strategy

import (
	http_entry "github.com/eolinker/apinto/entries/http-entry"
	"sort"
	"sync"
	"time"

	"github.com/eolinker/apinto/resources"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

var (
	actuatorSet ActuatorSet
)

const (
	fuseStatusTime = time.Minute * 30
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
		if next != nil {
			return next.DoChain(ctx)
		}
		return err
	}
	a.lock.RLock()
	handlers := a.handlers
	a.lock.RUnlock()
	entry := http_entry.NewEntry(httpCtx)
	var metrics string
	var fuseHandler *FuseHandler
	for _, h := range handlers {
		//check筛选条件
		if !h.filter.Check(httpCtx) {
			continue
		}
		fuseHandler = h
		break
	}
	if fuseHandler == nil {
		if next != nil {
			return next.DoChain(ctx)
		}
		return nil
	}
	metrics = fuseHandler.rule.metric.Metrics(entry)
	status := checkFuseStatus(ctx.Context(), metrics, cache)
	switch status {
	case fuseStatusFusing:
		fuseHandler.rule.response.Response(httpCtx)
		ctx.WithValue("is_block", true)
		ctx.SetLabel("block_name", fuseHandler.name)
		ctx.SetLabel("handler", "fuse")
		httpCtx.Response().SetHeader("Strategy-Fuse", fuseHandler.name)
		return nil
	default:
		if next != nil {
			err = next.DoChain(ctx)
			if err != nil {
				return err
			}
		}
		fuseHandler.Do(httpCtx, cache, metrics)
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
