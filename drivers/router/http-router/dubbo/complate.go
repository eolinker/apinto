package dubbo

import (
	"errors"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
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
	_, err := http_service.DubboAssert(org)
	if err != nil {
		return err
	}

	//TODO 调用http to dubbo sdk

	return nil
}
