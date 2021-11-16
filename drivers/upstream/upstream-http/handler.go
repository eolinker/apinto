package upstream_http

import (
	"fmt"
	"time"

	http_service "github.com/eolinker/eosc/http-service"
	"github.com/eolinker/goku/plugin"
)

var _ http_service.IChain = (*UpstreamHandler)(nil)

type UpstreamHandler struct {
	id            string
	upstrem       *Upstream
	retry         int
	timeout       time.Duration
	pluginsSource map[string]*plugin.Config
	orgFilter     plugin.IPlugin
}

func (u *UpstreamHandler) Destroy() {

	if u.orgFilter != nil {
		u.orgFilter.Destroy()
		u.orgFilter = nil
	}

}

func NewUpstreamHandler(id string, upstream *Upstream, retry int, timeout time.Duration, pluginsSource map[string]*plugin.Config) *UpstreamHandler {
	uh := &UpstreamHandler{
		id:            id,
		upstrem:       upstream,
		retry:         retry,
		timeout:       timeout,
		pluginsSource: pluginsSource,
		orgFilter:     nil,
	}
	uh.reset()
	return uh
}

func (u *UpstreamHandler) reset() {

	configs := u.upstrem.pluginConfig(u.pluginsSource)

	iPlugin := pluginManager.CreateUpstream(u.id, configs)

	u.orgFilter = iPlugin
}

//Send 请求发送
func (u *UpstreamHandler) DoChain(ctx http_service.IHttpContext) error {

	var lastErr error
	for doTrice := u.retry + 1; doTrice > 0; doTrice-- {

		node, err := u.upstrem.handler.Next()
		if err != nil {
			return err
		}
		scheme := node.Scheme()
		if scheme != "http" && scheme != "https" {
			scheme = u.upstrem.scheme
		}

		addr := fmt.Sprintf("%s://%s", scheme, node.Addr())
		filterSender := NewSendAddr(addr, u.timeout)
		if u.orgFilter == nil {
			lastErr = filterSender.DoFilter(ctx, nil)
		} else {
			lastErr = u.orgFilter.Append(filterSender).DoChain(ctx)
		}

	}

	return lastErr
}

type SendAddr struct {
	timeout time.Duration
	addr    string
}

func (s *SendAddr) Destroy() {

}

func NewSendAddr(addr string, timeout time.Duration) *SendAddr {
	return &SendAddr{timeout: timeout, addr: addr}
}

func (s *SendAddr) DoFilter(ctx http_service.IHttpContext, next http_service.IChain) (err error) {
	return ctx.SendTo(s.addr, s.timeout)
}
