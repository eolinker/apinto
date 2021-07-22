package http_proxy

import (
	"fmt"
	"net/url"
	"time"

	"github.com/eolinker/goku-eosc/node/http-proxy/backend"

	http_context "github.com/eolinker/goku-eosc/node/http-context"
	http_proxy_request "github.com/eolinker/goku-eosc/node/http-proxy/http-proxy-request"
)

//DoRequest 构造请求
func DoRequest(ctx *http_context.Context, uri string, timeout time.Duration) (backend.IResponse, error) {
	if uri == "" {
		return nil, fmt.Errorf("invaild url")
	}

	u, err := url.ParseRequestURI(uri)
	if err != nil {
		return nil, err
	}

	req, err := http_proxy_request.NewRequest(ctx.ProxyRequest.Method, u)
	if err != nil {

		return nil, err
	}

	queryDest := u.Query()
	if ctx.ProxyRequest.Queries() != nil {
		for k, vs := range ctx.ProxyRequest.Queries() {
			for _, v := range vs {
				queryDest.Add(k, v)
			}
		}
	}

	req.SetHeaders(ctx.ProxyRequest.Headers())

	req.SetQueryParams(queryDest)
	body, _ := ctx.ProxyRequest.RawBody()
	req.SetRawBody(body)
	if timeout != 0 {
		req.SetTimeout(timeout * time.Millisecond)
	}
	err = req.ParseBody()
	if err != nil {
		return nil, err
	}
	response, err := req.Send(ctx)
	if err != nil {
		return nil, err
	}

	return NewResponse(response)
}
