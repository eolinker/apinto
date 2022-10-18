package limiting_strategy

import (
	"github.com/eolinker/apinto/drivers/strategy/limiting-strategy/scalar"
	"github.com/eolinker/eosc/eocontext"
)

type ActuatorsHandler interface {
	Assert(ctx eocontext.EoContext) bool
	Check(ctx eocontext.EoContext, handlers []*LimitingHandler, queryScalar scalar.Manager, traffics scalar.Manager) error
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
