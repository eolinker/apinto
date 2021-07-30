package upstream_http_anonymous

import (
	"fmt"

	"github.com/eolinker/goku-eosc/node/http-proxy/backend"

	"github.com/eolinker/goku-eosc/service"

	http_proxy "github.com/eolinker/goku-eosc/node/http-proxy"

	http_context "github.com/eolinker/goku-eosc/node/http-context"

	"github.com/eolinker/goku-eosc/utils"
)

func NewAnonymousUpstream() *httpUpstream {
	return &httpUpstream{}
}

//Http org
type httpUpstream struct {
}

//send 请求发送，忽略重试
func (h *httpUpstream) Send(ctx *http_context.Context, serviceDetail service.IServiceDetail) (backend.IResponse, error) {
	var response backend.IResponse
	var err error
	path := utils.TrimPrefixAll(ctx.ProxyRequest.TargetURL(), "/")
	for doTrice := serviceDetail.Retry() + 1; doTrice > 0; doTrice-- {
		u := fmt.Sprintf("%s://%s/%s", serviceDetail.Scheme(), serviceDetail.ProxyAddr(), path)
		response, err = http_proxy.DoRequest(ctx, u, serviceDetail.Timeout())

		if err != nil {
			continue
		} else {
			return response, err
		}
	}

	return response, err
}
