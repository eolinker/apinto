package upstream_http

import (
	"fmt"
	"time"

	"github.com/eolinker/goku/plugin"
	"github.com/eolinker/goku/upstream"

	http_service "github.com/eolinker/eosc/http-service"
	"github.com/eolinker/goku/discovery"
	"github.com/eolinker/goku/upstream/balance"
)

var (
	_ upstream.IUpstreamCreate = (*Upstream)(nil)
)

type Upstream struct {
	scheme  string
	app     discovery.IApp
	handler balance.IBalanceHandler
	retry   int
	timeout time.Duration
}

func (up *Upstream) Create(id string, configs map[string]*plugin.Config) upstream.IUpstream {
	panic("implement me")
}

func NewUpstream(scheme string, app discovery.IApp, handler balance.IBalanceHandler) *Upstream {
	return &Upstream{scheme: scheme, app: app, handler: handler}
}

//Send 请求发送，忽略重试
func (up *Upstream) Send(ctx http_service.IHttpContext) error {

	var lastErr error
	for doTrice := up.retry + 1; doTrice > 0; doTrice-- {

		node, err := up.handler.Next()
		if err != nil {
			return err
		}
		scheme := node.Scheme()
		if scheme != "http" && scheme != "https" {
			scheme = up.scheme
		}

		addr := fmt.Sprintf("%s://%s", scheme, node.Addr())
		lastErr = ctx.SendTo(addr, up.timeout)
		if lastErr != nil {
			node.Down()
			//处理不可用节点
			up.app.NodeError(node.ID())

			continue
		}
	}

	return lastErr
}
