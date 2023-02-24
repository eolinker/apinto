package dubbo2_router

import (
	"github.com/eolinker/eosc/eocontext"
)

type finishHandler struct {
}

func newFinishHandler() *finishHandler {
	return &finishHandler{}
}

func (f *finishHandler) Finish(org eocontext.EoContext) error {

	return nil
}
