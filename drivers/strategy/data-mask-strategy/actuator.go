package data_mask_strategy

import (
	"sort"
	"sync"
	
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	
	"github.com/eolinker/apinto/strategy"
	"github.com/eolinker/eosc/eocontext"
)

var (
	actuatorSet ActuatorSet
)

func init() {
	actuator := newtActuator()
	actuatorSet = actuator
}

type ActuatorSet interface {
	strategy.IStrategyHandler
	Set(string, *handler)
	Del(id string)
}

type tActuator struct {
	lock     sync.RWMutex
	all      map[string]*handler
	handlers []*handler
}

func (a *tActuator) Destroy() {

}

func (a *tActuator) Set(id string, val *handler) {
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
	
	handlers := make([]*handler, 0, len(a.all))
	for _, h := range a.all {
		handlers = append(handlers, h)
	}
	sort.Slice(handlers, func(i, j int) bool {
		return handlers[i].priority < handlers[j].priority
	})
	a.lock.Lock()
	defer a.lock.Unlock()
	a.handlers = handlers
}
func newtActuator() *tActuator {
	return &tActuator{
		all: make(map[string]*handler),
	}
}

func (a *tActuator) Strategy(ctx eocontext.EoContext, next eocontext.IChain) error {
	
	httpCtx, err := http_service.Assert(ctx)
	if err != nil {
		return err
	}
	if next != nil {
		err = next.DoChain(ctx)
		if err != nil {
			return err
		}
	}
	a.lock.RLock()
	handlers := a.handlers
	a.lock.RUnlock()
	//var execHandler *handler
	for _, h := range handlers {
		// 匹配Filter
		if !h.filter.Check(httpCtx) {
			// 未命中，下一条规则
			continue
		}
		err = h.ResponseExec(httpCtx)
		if err != nil {
			return err
		}
		ctx.WithValue("is_block", true)
		ctx.SetLabel("block_name", h.name)
		ctx.SetLabel("handler", "data_mask")
		//execHandler = h
		// 匹配中后，跳出循环
		break
	}
	
	return nil
}

func DoStrategy(ctx eocontext.EoContext, next eocontext.IChain) error {
	return actuatorSet.Strategy(ctx, next)
}
