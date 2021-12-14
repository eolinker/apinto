package http_context

import (
	"strconv"
	"strings"

	http_service "github.com/eolinker/eosc/http-service"

	"github.com/valyala/fasthttp"
)

var _ http_service.IResponse = (*Response)(nil)

type Response struct {
	*ResponseHeader
	*fasthttp.Response
	responseError error
}

func (r *Response) ResponseError() error {
	return r.responseError
}

func (r *Response) ClearError() {
	r.responseError = nil
}

func (r *Response) reset() error {
	r.ResponseHeader.tmp = nil
	return nil
}

func NewResponse(ctx *fasthttp.RequestCtx) *Response {
	return &Response{Response: &ctx.Response, ResponseHeader: NewResponseHeader(&ctx.Response.Header)}
}

func (r *Response) BodyLen() int {
	return r.header.ContentLength()
}

func (r *Response) GetBody() []byte {
	if strings.Contains(r.GetHeader("Content-Encoding"), "gzip") {
		body, _ := r.BodyGunzip()
		r.Headers().Del("Content-Encoding")
		r.SetHeader("Content-Length", strconv.Itoa(len(body)))
		r.Response.SetBody(body)
	}

	return r.Response.Body()
}

func (r *Response) StatusCode() int {
	if r.responseError != nil {
		return 504
	}
	return r.Response.StatusCode()
}

func (r *Response) Status() string {
	return strconv.Itoa(r.StatusCode())
}

func (r *Response) SetStatus(code int, status string) {
	r.Response.SetStatusCode(code)
}

func (r *Response) SetBody(bytes []byte) {
	r.Response.SetBody(bytes)
}
