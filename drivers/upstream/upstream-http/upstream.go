package upstream_http

import (
	"fmt"

	http_service "github.com/eolinker/eosc/http-service"
	"github.com/eolinker/goku/discovery"
	http_proxy "github.com/eolinker/goku/node/http-proxy"
	"github.com/eolinker/goku/upstream/balance"
	"github.com/eolinker/goku/utils"
	"github.com/valyala/fasthttp"
)

type Upstream struct {
	scheme  string
	app     discovery.IApp
	handler balance.IBalanceHandler
}

func NewUpstream(scheme string, app discovery.IApp, handler balance.IBalanceHandler) *Upstream {
	return &Upstream{scheme: scheme, app: app, handler: handler}
}

//Send 请求发送，忽略重试
func (h *Upstream) Send(ctx http_service.IHttpContext) (*fasthttp.Response, error) {
	var response *fasthttp.Response
	var err error

	path := utils.TrimPrefixAll(uri, "/")
	request := ctx.Proxy()
	for doTrice := serviceDetail.Retry() + 1; doTrice > 0; doTrice-- {
		var node discovery.INode
		node, err = h.handler.Next()
		if err != nil {
			return nil, err
		}
		scheme := node.Scheme()
		if scheme != "http" && scheme != "https" {
			scheme = h.scheme
		}
		request.u(fmt.Sprintf("%s://%s/%s", scheme, node.Addr(), path))
		response, err = http_proxy.DoRequest(request, serviceDetail.Timeout())

		if err != nil {
			if response == nil {
				node.Down()
			}
			//处理不可用节点
			h.app.NodeError(node.ID())
			continue
		} else {
			return response, nil
		}
	}

	return response, err
}
