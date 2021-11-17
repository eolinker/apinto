package http_context

import (
	"strconv"

	http_service "github.com/eolinker/eosc/http-service"

	"github.com/valyala/fasthttp"
)

var _ http_service.IResponse = (*Response)(nil)

type Response struct {
	*ResponseHeader
	*fasthttp.Response
}

func (r *Response) reset() error {
	r.ResponseHeader.tmp = nil
	return nil
}

func NewResponse(ctx *fasthttp.RequestCtx) *Response {
	return &Response{Response: &ctx.Response, ResponseHeader: NewResponseHeader(&ctx.Response.Header)}
}

func (r *Response) GetBody() []byte {
	return r.Response.Body()
}

func (r *Response) StatusCode() int {
	return r.Response.StatusCode()
}

func (r *Response) Status() string {
	return strconv.Itoa(r.Response.StatusCode())
}

func (r *Response) SetStatus(code int, status string) {
	r.Response.SetStatusCode(code)
}

func (r *Response) SetBody(bytes []byte) {
	r.Response.SetBody(bytes)
}
