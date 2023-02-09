package grpc_router

import (
	"errors"
	"time"

	"github.com/eolinker/eosc/eocontext"
)

var (
	ErrorTimeoutComplete = errors.New("complete timeout")
)

type Complete struct {
	retry   int
	timeOut time.Duration
}

func NewComplete(retry int, timeOut time.Duration) *Complete {
	return &Complete{retry: retry, timeOut: timeOut}
}

func (h *Complete) Complete(org eocontext.EoContext) error {
	panic("need finish")
}

type CompleteCaller struct {
}

func NewCompleteCaller() *CompleteCaller {
	return &CompleteCaller{}
}

func (h *CompleteCaller) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return ctx.GetComplete().Complete(ctx)
}

func (h *CompleteCaller) Destroy() {

}
