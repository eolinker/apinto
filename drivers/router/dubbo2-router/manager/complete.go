package manager

import (
	"errors"
	"time"

	"github.com/eolinker/eosc/eocontext"
	dubbo2_context "github.com/eolinker/eosc/eocontext/dubbo2-context"
	"github.com/eolinker/eosc/log"
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
	ctx, err := dubbo2_context.Assert(org)
	if err != nil {
		return err
	}

	//设置响应开始时间
	proxyTime := time.Now()
	defer func() {
		ctx.Response().SetResponseTime(time.Now().Sub(proxyTime))
	}()

	balance := ctx.GetBalance()
	var lastErr error

	timeOut := balance.TimeOut()
	for index := 0; index <= h.retry; index++ {

		if h.timeOut > 0 && time.Now().Sub(proxyTime) > h.timeOut {
			ctx.Response().SetBody(Dubbo2ErrorResult(ErrorTimeoutComplete))
			return ErrorTimeoutComplete
		}
		node, _, err := balance.Select(ctx)
		if err != nil {
			log.Error("select error: ", err)
			ctx.Response().SetBody(Dubbo2ErrorResult(errors.New("node is null")))
			return err
		}

		lastErr = ctx.Invoke(node, timeOut)
		if lastErr == nil {
			return nil
		}
		log.Error("dubbo upstream send error: ", lastErr)
	}

	ctx.Response().SetBody(Dubbo2ErrorResult(lastErr))

	return lastErr
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
