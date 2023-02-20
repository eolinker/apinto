package http_to_dubbo2

import (
	"encoding/json"
	"errors"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
	"time"
)

var (
	ErrorTimeoutComplete = errors.New("complete timeout")
)

type Complete struct {
	retry        int
	timeOut      time.Duration
	dubbo2Client *dubbo2Client
}

func NewComplete(retry int, timeOut time.Duration, dubbo2Client *dubbo2Client) *Complete {
	return &Complete{retry: retry, timeOut: timeOut, dubbo2Client: dubbo2Client}
}

func (c *Complete) Complete(org eocontext.EoContext) error {
	ctx, err := http_service.Assert(org)
	if err != nil {
		return err
	}
	//设置响应开始时间
	proxyTime := time.Now()

	defer func() {
		//设置原始响应状态码
		ctx.Response().SetProxyStatus(ctx.Response().StatusCode(), "")
		//设置上游响应总时间, 单位为毫秒
		//ctx.WithValue("response_time", time.Now().Sub(proxyTime).Milliseconds())
		ctx.Response().SetResponseTime(time.Now().Sub(proxyTime))
		ctx.SetLabel("handler", "proxy")
	}()

	balance := ctx.GetBalance()
	var lastErr error

	for index := 0; index <= c.retry; index++ {

		if c.timeOut > 0 && time.Now().Sub(proxyTime) > c.timeOut {
			return ErrorTimeoutComplete
		}
		node, err := balance.Select(ctx)
		if err != nil {
			log.Error("select error: ", err)
			ctx.Response().SetStatus(501, "501")
			ctx.Response().SetBody([]byte(err.Error()))
			return err
		}

		log.Debug("node: ", node.Addr())
		var result interface{}
		var dialErr error
		result, dialErr = c.dubbo2Client.dial(ctx.Context(), node.Addr(), c.timeOut)
		lastErr = dialErr
		if lastErr == nil {
			bytes, err := json.Marshal(result)
			if err != nil {
				ctx.Response().SetBody([]byte(err.Error()))
				return nil
			}
			ctx.Response().SetBody(bytes)
			return nil
		}
		log.Error("http to dubbo2 dial error: ", lastErr)
	}

	ctx.Response().SetBody([]byte(lastErr.Error()))

	return lastErr
}
