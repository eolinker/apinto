package visit_strategy

import (
	"github.com/eolinker/apinto/strategy"
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
	strategy.IStrategyHandler
	Set(string, *visitHandler)
	Del(id string)
}

type tActuator struct {
	lock     sync.RWMutex
	all      map[string]*visitHandler
	handlers []*visitHandler
}

func (a *tActuator) Destroy() {

}

func (a *tActuator) Set(id string, val *visitHandler) {
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

	handlers := make([]*visitHandler, 0, len(a.all))
	for _, h := range a.all {
		if !h.stop {
			handlers = append(handlers, h)
		}
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
		all: make(map[string]*visitHandler),
	}
}

func (a *tActuator) Strategy(ctx eocontext.EoContext, next eocontext.IChain) error {

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

		//第一个判断条件为访问规则必须是允许,并且生效范围检测出是黑名单                 第二个判断条件为访问规则必须是拒绝,并且生效返回检测出是黑名单
		if (handler.rule.visit && !handler.rule.effectFilter.Check(ctx)) || (!handler.rule.visit && handler.rule.effectFilter.Check(ctx)) {
			httpCtx.Response().SetStatus(403, "")
			return nil
		}

		if handler.rule.isContinue {
			continue
		}
		break

	}

	if next != nil {
		return next.DoChain(ctx)
	}
	return nil
}

func DoStrategy(ctx eocontext.EoContext, next eocontext.IChain) error {
	return actuatorSet.Strategy(ctx, next)
}
