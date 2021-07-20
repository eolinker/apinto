package upstream_http

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/eolinker/goku-eosc/service"

	"github.com/eolinker/goku-eosc/upstream/balance"

	http_proxy "github.com/eolinker/goku-eosc/upstream/upstream-http/http-proxy"

	http_context "github.com/eolinker/goku-eosc/node/http-context"

	"github.com/eolinker/goku-eosc/utils"
)

//Http org
type httpUpstream struct {
	Scheme string  `json:"scheme"`
	Nodes  []*node `json:"nodes" yaml:"nodes"`
	Type   string  `json:"type"`
}

type node struct {
	IP     string            `json:"ip" yaml:"ip"`
	Port   int               `json:"port" yaml:"port"`
	Labels map[string]string `json:"labels" yaml:"labels"`
}

//send 请求发送，忽略重试
func (h *httpUpstream) Send(ctx *http_context.Context, serviceDetail service.IServiceDetail, handler balance.IBalanceHandler) (*http.Response, error) {
	var response *http.Response
	var err error
	path := utils.TrimPrefixAll(ctx.ProxyRequest.TargetURL(), "/")
	node, err := handler.Next()
	if err != nil {
		return nil, err
	}
	for doTrice := serviceDetail.GetRetry() + 1; doTrice > 0; doTrice-- {

		u := fmt.Sprintf("%s://%s/%s", h.Scheme, node.Addr(), path)
		response, err = http_proxy.DoRequest(ctx, u, serviceDetail.GetTimeout())

		if err != nil {
			node, err = handler.Next()
			if err != nil {
				return nil, err
			}
			continue
		} else {
			return response, err
		}
	}

	return response, err
}

func GetType() reflect.Type {
	return reflect.TypeOf((*httpUpstream)(nil))
}
