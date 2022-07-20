package test

import (
	"net/http"
	"strconv"

	http_service "github.com/eolinker/eosc/context/http-context"
)

var _ http_service.IResponse = (*Response)(nil)

type Response struct {
	*ResponseHeader
	body       []byte
	statusCode int
}

func (r *Response) reset() error {
	r.ResponseHeader.header = make(http.Header)
	return nil
}

func NewResponse() *Response {
	return &Response{ResponseHeader: NewResponseHeader()}
}

func (r *Response) GetBody() []byte {
	return r.body
}

func (r *Response) StatusCode() int {
	return r.statusCode
}

func (r *Response) Status() string {
	return strconv.Itoa(r.statusCode)

}

func (r *Response) SetStatus(code int, status string) {
	r.statusCode = code
}

func (r *Response) SetBody(bytes []byte) {
	r.body = bytes
}
