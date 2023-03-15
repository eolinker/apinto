package http_complete

import (
	"errors"
	"fmt"
	"github.com/eolinker/apinto/entries/ctx_key"
	"github.com/eolinker/apinto/entries/router"
	"strconv"
	"strings"
	"time"

	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
)

var (
	ErrorTimeoutComplete = errors.New("complete timeout")
)

type HttpComplete struct {
}

func NewHttpComplete() *HttpComplete {
	return &HttpComplete{}
}

func (h *HttpComplete) Complete(org eocontext.EoContext) error {

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
	app := ctx.GetApp()
	var lastErr error
	scheme := app.Scheme()

	switch strings.ToLower(scheme) {
	case "", "tcp":
		scheme = "http"
	case "tsl", "ssl", "https":
		scheme = "https"

	}
	timeOut := app.TimeOut()

	retryValue := ctx.Value(ctx_key.CtxKeyRetry)
	retry, ok := retryValue.(int)
	if !ok {
		retry = router.DefaultRetry
	}

	timeoutValue := ctx.Value(ctx_key.CtxKeyTimeout)
	timeout, ok := timeoutValue.(time.Duration)
	if !ok {
		timeout = router.DefaultTimeout
	}

	for index := 0; index <= retry; index++ {

		if timeout > 0 && time.Now().Sub(proxyTime) > timeout {
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
		addr := fmt.Sprintf("%s://%s", scheme, node.Addr())
		lastErr = ctx.SendTo(addr, timeOut)
		if lastErr == nil {
			return nil
		}
		log.Error("http upstream send error: ", lastErr)
	}

	return lastErr
}

type NoServiceCompleteHandler struct {
	status int
	header map[string]string
	body   string
}

func NewNoServiceCompleteHandler(status int, header map[string]string, body string) *NoServiceCompleteHandler {
	return &NoServiceCompleteHandler{status: status, header: header, body: body}
}

func (n *NoServiceCompleteHandler) Complete(org eocontext.EoContext) error {
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
	for key, value := range n.header {
		ctx.Response().SetHeader(key, value)
	}
	ctx.Response().SetBody([]byte(n.body))
	ctx.Response().SetStatus(n.status, strconv.Itoa(n.status))
	return nil
}

type httpCompleteCaller struct {
}

func NewHttpCompleteCaller() *httpCompleteCaller {
	return &httpCompleteCaller{}
}

func (h *httpCompleteCaller) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return ctx.GetComplete().Complete(ctx)
}

func (h *httpCompleteCaller) Destroy() {

}
