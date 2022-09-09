package limiting_stragety

import (
	"github.com/eolinker/apinto/strategy"
	"github.com/eolinker/eosc/eocontext"
)

var (
	actuatorSet     ActuatorSet
	actuatorHandler strategy.IStrategyHandler
)

func init() {
	actuator := newActuator()
	actuatorSet = actuator

	strategy.AddStrategyHandler(actuatorHandler)
}

type ActuatorSet interface {
	Set(id string, limiting *Limiting)
	Del(id string)
}

type tActuator struct {
}

func (a *tActuator) Set(id string, limiting *Limiting) {

}

func (a *tActuator) Del(id string) {

}

func newActuator() *tActuator {
	return &tActuator{}
}

func (a *tActuator) Strategy(ctx eocontext.EoContext, next eocontext.IChain) error {
	//TODO implement me
	panic("implement me")
}
