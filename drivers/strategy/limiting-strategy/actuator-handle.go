package limiting_strategy

import (
	"github.com/eolinker/eosc/eocontext"
)

type ActuatorsHandler interface {
	Assert(ctx eocontext.EoContext) bool
	Check(ctx eocontext.EoContext, handlers []*LimitingHandler, scalars *Scalars) error
}

var (
	actuatorsHandlers []ActuatorsHandler
)

func RegisterActuator(handler ActuatorsHandler) {
	
	actuatorsHandlers = append(actuatorsHandlers, handler)
}
func getActuatorsHandlers() []ActuatorsHandler {
	
	return actuatorsHandlers
}
