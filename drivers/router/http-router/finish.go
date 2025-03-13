package http_router

import (
	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
)

var defaultFinisher = &Finisher{}

type Finisher struct {
}

func (f *Finisher) Finish(org eocontext.EoContext) error {
	ctx, err := http_context.Assert(org)
	if err != nil {
		return err
	}
	ctx.SetLabel("current_running", "false")
	ctx.FastFinish()

	return nil
}
