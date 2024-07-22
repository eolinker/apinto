package visit_strategy

import (
	"errors"
	"sort"
	"sync"

	"github.com/eolinker/apinto/strategy"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
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
		if next != nil {
			return next.DoChain(ctx)
		}
		return err
	}

	a.lock.RLock()
	handlers := a.handlers
	a.lock.RUnlock()
	pass := true
	var name string
	for _, handler := range handlers {
		// 匹配Filter
		if !handler.filter.Check(httpCtx) {
			// 未命中，下一条规则
			continue
		}
		// 匹配资源
		match := handler.rule.effectFilter.Check(ctx)
		if match {
			// 匹配成功
			pass = handler.rule.visit
			name = handler.name
			break
		}
		pass = !handler.rule.visit
		name = handler.name
		if handler.rule.isContinue {
			continue
		}
		break
	}
	if !pass {
		ctx.SetLabel("handler", "visit")
		httpCtx.Response().SetStatus(403, "")
		errInfo := "not allowed"
		httpCtx.Response().SetBody([]byte(errInfo))
		ctx.WithValue("is_block", true)
		ctx.SetLabel("block_name", name)
		return errors.New(errInfo)
	}
	if next != nil {
		return next.DoChain(ctx)
	}
	return nil
}

func DoStrategy(ctx eocontext.EoContext, next eocontext.IChain) error {
	return actuatorSet.Strategy(ctx, next)
}
