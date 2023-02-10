package dubbo_router

import (
	"errors"
	"github.com/eolinker/eosc/eocontext"
	dubbo_context "github.com/eolinker/eosc/eocontext/dubbo-context"
	"github.com/eolinker/eosc/log"
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
	ctx, err := dubbo_context.Assert(org)
	if err != nil {
		return err
	}

	//设置响应开始时间
	proxyTime := time.Now()
	balance := ctx.GetBalance()
	app := ctx.GetApp()
	var lastErr error

	timeOut := app.TimeOut()
	for index := 0; index <= h.retry; index++ {

		if h.timeOut > 0 && time.Now().Sub(proxyTime) > h.timeOut {
			return ErrorTimeoutComplete
		}
		node, err := balance.Select(ctx)
		if err != nil {
			log.Error("select error: ", err)
			//ctx.Response().SetStatus(501, "501")
			//ctx.Response().SetBody([]byte(err.Error()))
			return err
		}

		log.Debug("node: ", node.Addr())
		lastErr = ctx.Invoke(node.Addr(), timeOut)
		if lastErr == nil {
			return nil
		}
		log.Error("dubbo upstream send error: ", lastErr)
	}

	return lastErr
}
