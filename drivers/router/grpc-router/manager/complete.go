package manager

import (
	"errors"
	"time"

	grpc_context "github.com/eolinker/eosc/eocontext/grpc-context"
	"github.com/eolinker/eosc/log"

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

	ctx, err := grpc_context.Assert(org)
	if err != nil {
		return err
	}
	var lastErr error

	//设置响应开始时间
	proxyTime := time.Now()

	balance := ctx.GetBalance()
	app := ctx.GetApp()
	if app.Scheme() == "https" {
		ctx.EnableTls(true)
	}
	defer func() {
		//设置上游响应总时间, 单位为毫秒
		ctx.Response().SetErr(lastErr)
		ctx.SetLabel("handler", "proxy")
	}()
	timeOut := app.TimeOut()
	for index := 0; index <= h.retry; index++ {

		if h.timeOut > 0 && time.Now().Sub(proxyTime) > h.timeOut {
			return ErrorTimeoutComplete
		}
		node, err := balance.Select(ctx)
		if err != nil {
			return err
		}

		lastErr = ctx.Invoke(node, timeOut)
		if lastErr == nil {
			return nil
		}
		log.Error("grpc upstream send error: ", lastErr)
	}

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
