package service_http

import (
	"fmt"
	"time"

	"github.com/eolinker/eosc/log"

	http_service "github.com/eolinker/eosc/http-service"
)

var _ http_service.IChain = (*UpstreamHandler)(nil)

type UpstreamHandler struct {
	id       string
	upstream *Upstream
	retry    int
	timeout  time.Duration
}

func (u *UpstreamHandler) Destroy() {

	upstream := u.upstream
	if upstream != nil {
		u.upstream = nil
		upstream.handlers.Del(u.id)
	}

}

func NewUpstreamHandler(id string, upstream *Upstream, retry int, timeout time.Duration) *UpstreamHandler {
	uh := &UpstreamHandler{
		id:       id,
		upstream: upstream,
		retry:    retry,
		timeout:  timeout,
	}
	return uh
}

//DoChain 请求发送
func (u *UpstreamHandler) DoChain(ctx http_service.IHttpContext) error {

	var lastErr error

	//设置响应开始时间
	proxyTime := time.Now()

	defer func() {
		//设置原始响应状态码
		ctx.Response().SetProxyStatus(ctx.Response().StatusCode(), "")
		//设置上游响应时间, 单位为毫秒
		ctx.WithValue("response_time", time.Now().Sub(proxyTime).Milliseconds())
	}()

	for doTrice := u.retry + 1; doTrice > 0; doTrice-- {

		node, err := u.upstream.handler.Next()
		if err != nil {
			return err
		}
		scheme := node.Scheme()
		if scheme != "http" && scheme != "https" {
			scheme = u.upstream.scheme
		}
		log.Debug("node: ", node.Addr())
		addr := fmt.Sprintf("%s://%s", scheme, node.Addr())
		lastErr = ctx.SendTo(addr, u.timeout)
		if lastErr == nil {
			return nil
		}
		log.Error("http upstream send error: ", lastErr)
	}

	return lastErr
}
