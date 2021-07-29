package backend

import "net/http"

//IResponse 响应接口
type IResponse interface {
	Body() []byte
	StatusCode() int
	Header() http.Header
	Proto() string
	Status() string
}
