package upstream_http

import (
	http_context "github.com/eolinker/goku/node/http-context"
	"github.com/valyala/fasthttp"
)

type UpstreamHandler struct {
}

func (u *UpstreamHandler) Send(ctx *http_context.Context) (*fasthttp.Response, error) {
	panic("implement me")
}
