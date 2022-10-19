package limiting_strategy

import (
	"github.com/eolinker/apinto/drivers/strategy/limiting-strategy/scalar"
	"github.com/eolinker/apinto/resources"
	"github.com/eolinker/eosc/eocontext"
	"sort"
	"sync"
)

var (
	actuatorSet ActuatorSet
)

func init() {
	actuator := newActuator()
	actuatorSet = actuator

}

type ActuatorSet interface {
	Strategy(ctx eocontext.EoContext, next eocontext.IChain, cache resources.ICache) error
	Set(id string, limiting *LimitingHandler)
	Del(id string)
}

type tActuatorSet struct {
	lock        sync.RWMutex
	all         map[string]*LimitingHandler
	handlers    []*LimitingHandler
	queryScalar scalar.Manager
	traffics    scalar.Manager
}

func (a *tActuatorSet) Destroy() {

}

func (a *tActuatorSet) Set(id string, limiting *LimitingHandler) {
	// 调用来源有锁
	a.all[id] = limiting
	a.rebuild()

}

func (a *tActuatorSet) Del(id string) {
	// 调用来源有锁
	delete(a.all, id)
	a.rebuild()
}

func (a *tActuatorSet) rebuild() {

	handlers := make([]*LimitingHandler, 0, len(a.all))
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
func newActuator() *tActuatorSet {
	return &tActuatorSet{
		queryScalar: scalar.NewManager(),
		traffics:    scalar.NewManager(),
		all:         make(map[string]*LimitingHandler),
	}
}

func (a *tActuatorSet) Strategy(ctx eocontext.EoContext, next eocontext.IChain, cache resources.ICache) error {

	a.lock.RLock()
	handlers := a.handlers
	a.lock.RUnlock()
	acs := getActuatorsHandlers()
	for _, ach := range acs {
		if ach.Assert(ctx) {
			err := ach.Check(ctx, handlers, a.queryScalar, a.traffics)
			if err != nil {
				return err
			}
			break
		}
	}

	if next != nil {
		return next.DoChain(ctx)
	}
	return nil
}

type handlerListSort []*LimitingHandler

func (hs handlerListSort) Len() int {
	return len(hs)
}

func (hs handlerListSort) Less(i, j int) bool {

	return hs[i].priority < hs[j].priority
}

func (hs handlerListSort) Swap(i, j int) {
	hs[i], hs[j] = hs[j], hs[i]
}

func DoStrategy(ctx eocontext.EoContext, next eocontext.IChain, cache resources.ICache) error {
	return actuatorSet.Strategy(ctx, next, cache)
}
