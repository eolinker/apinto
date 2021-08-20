package httplog_transporter

import (
	"github.com/eolinker/eosc/log"
	"net/http"
)

type Config struct {
	Method  string
	Url     string
	Headers http.Header
	Level   log.Level

	HandlerCount int
}
