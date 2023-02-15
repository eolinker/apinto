package dubbo2_router

import (
	"github.com/eolinker/eosc/eocontext"
	dubbo2_context "github.com/eolinker/eosc/eocontext/dubbo2-context"
)

type finishHandler struct {
}

func newFinishHandler() *finishHandler {
	return &finishHandler{}
}

func (f *finishHandler) Finish(org eocontext.EoContext) error {
	_, err := dubbo2_context.Assert(org)
	if err != nil {
		return err
	}
	//todo write

	return nil
}
