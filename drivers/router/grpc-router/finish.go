package grpc_router

import (
	"github.com/eolinker/eosc/eocontext"
	grpc_context "github.com/eolinker/eosc/eocontext/grpc-context"
)

var defaultFinisher = &Finisher{}

type Finisher struct {
}

func (f *Finisher) Finish(org eocontext.EoContext) error {
	ctx, err := grpc_context.Assert(org)
	if err != nil {
		return err
	}
	ctx.Response()

	return nil
}
