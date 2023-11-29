package limiting_strategy

import (
	"sort"
	"sync"

	"github.com/eolinker/eosc/eocontext"

	"github.com/eolinker/apinto/resources"
)

var (
	actuatorSet ActuatorSet
)

func init() {
	actuator := newActuator()
	actuatorSet = actuator
}

type ActuatorSet interface {
	Strategy(ctx eocontext.EoContext, next eocontext.IChain, scalars *Scalars) error
	Set(id string, limiting *LimitingHandler)
	Del(id string)
}

type tActuatorSet struct {
	lock     sync.RWMutex
	all      map[string]*LimitingHandler
	handlers []*LimitingHandler
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
		all: make(map[string]*LimitingHandler),
	}
}

func (a *tActuatorSet) Strategy(ctx eocontext.EoContext, next eocontext.IChain, scalars *Scalars) error {
	a.lock.RLock()
	handlers := a.handlers
	a.lock.RUnlock()
	acs := getActuatorsHandlers()
	for _, ach := range acs {
		if ach.Assert(ctx) {
			err := ach.Check(ctx, handlers, scalars)
			if err != nil {
				ctx.SetLabel("handler", "limiting")
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

type Scalars struct {
	QuerySecond resources.Vector
	QueryMinute resources.Vector
	QueryHour   resources.Vector

	TrafficsSecond resources.Vector
	TrafficsMinute resources.Vector
	TrafficsHour   resources.Vector
}

func DoStrategy(ctx eocontext.EoContext, next eocontext.IChain, scalars *Scalars) error {
	return actuatorSet.Strategy(ctx, next, scalars)
}
