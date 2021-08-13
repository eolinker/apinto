package upstream_http_anonymous

import (
	"fmt"

	"github.com/valyala/fasthttp"

	"github.com/eolinker/goku/service"

	http_proxy "github.com/eolinker/goku/node/http-proxy"

	http_context "github.com/eolinker/goku/node/http-context"

	"github.com/eolinker/goku/utils"
)

func NewAnonymousUpstream() *httpUpstream {
	return &httpUpstream{}
}

//Http org
type httpUpstream struct {
}

//send 请求发送，忽略重试
func (h *httpUpstream) Send(ctx *http_context.Context, serviceDetail service.IServiceDetail) (*fasthttp.Response, error) {
	var response *fasthttp.Response
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
