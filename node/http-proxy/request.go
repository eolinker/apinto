package http_proxy

import (
	"time"

	"github.com/valyala/fasthttp"

	http_proxy_request "github.com/eolinker/goku/node/http-proxy/http-proxy-request"
)

//DoRequest 构造请求
func DoRequest(req *fasthttp.Request, timeout time.Duration) (*fasthttp.Response, error) {
	newReq := http_proxy_request.NewRequest(req, timeout)
	response, err := newReq.Send()
	if err != nil {
		return nil, err
	}

	return response, nil
}
