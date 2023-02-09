package dubbo_router

import (
	"errors"
	"github.com/eolinker/eosc/eocontext"
	dubbo_context "github.com/eolinker/eosc/eocontext/dubbo-context"
	"time"
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
	_, err := dubbo_context.Assert(org)
	if err != nil {
		return err
	}

	//TODO 调用http to dubbo sdk

	return nil
}
