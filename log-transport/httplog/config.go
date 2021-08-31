package httplog

import (
	"net/http"

	"github.com/eolinker/eosc/log"
)

type Config struct {
	Method  string
	Url     string
	Headers http.Header
	Level   log.Level

	HandlerCount int
}
