package cache_strategy

import (
	"github.com/eolinker/eosc/eocontext"
)

type ActuatorsHandler interface {
	Assert(ctx eocontext.EoContext) bool
	Check(ctx eocontext.EoContext, handlers []*CacheValidTimeHandler) error
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
