package limiting_stragety

import (
	"github.com/eolinker/apinto/drivers/strategy/limiting-stragety/http"
	"github.com/eolinker/apinto/drivers/strategy/limiting-stragety/scalar"
	"github.com/eolinker/eosc/eocontext"
)

func init() {
	registerActuator(http.NewActuator())
}

type ActuatorsHandler interface {
	Assert(ctx eocontext.EoContext) bool
	Check(ctx eocontext.EoContext, handlers []*LimitingHandler, queryScalar scalar.Manager, traffics scalar.Manager) error
}

var (
	actuatorsHandlers []ActuatorsHandler
)

func registerActuator(handler ActuatorsHandler) {

	actuatorsHandlers = append(actuatorsHandlers, handler)
}
func getActuatorsHandlers() []ActuatorsHandler {

	return actuatorsHandlers
}
