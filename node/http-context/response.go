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

func (r *Response) Finish() error {

	fasthttp.ReleaseResponse(r.Response)
	return nil
}

func NewResponse() *Response {
	response := fasthttp.AcquireResponse()
	return &Response{Response: fasthttp.AcquireResponse(), ResponseHeader: NewResponseHeader(&response.Header)}
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
func (r *Response) Set(response *fasthttp.Response) {
	if response != nil {
		r.Response.Reset()
		response.CopyTo(r.Response)
		r.ResponseHeader = NewResponseHeader(&r.Response.Header)
	}

}
