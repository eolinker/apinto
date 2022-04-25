package http_transport

import "net/http"

type Config struct {
	Method       string
	Url          string
	Headers      http.Header
	HandlerCount int
}
