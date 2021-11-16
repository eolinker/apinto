package http_context

import (
	"strconv"

	http_service "github.com/eolinker/eosc/http-service"

	"github.com/valyala/fasthttp"
)

var _ http_service.IResponse = (*Response)(nil)

type Response struct {
	ResponseHeader
	response *fasthttp.Response
}

func NewResponse(response *fasthttp.Response) *Response {
	return &Response{response: response, ResponseHeader: ResponseHeader{
		header: &response.Header,
	}}
}

func (r *Response) GetBody() []byte {
	return r.response.Body()
}

func (r *Response) StatusCode() int {
	return r.response.StatusCode()
}

func (r *Response) Status() string {
	return strconv.Itoa(r.response.StatusCode())
}

func (r *Response) SetStatus(code int, status string) {
	r.response.SetStatusCode(code)
}

func (r *Response) SetBody(bytes []byte) {
	r.response.SetBody(bytes)
}
