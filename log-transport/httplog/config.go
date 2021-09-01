package httplog

import (
	"net/http"

	"github.com/eolinker/eosc/log"
)

//Config httplog-Transporter所需配置
type Config struct {
	Method  string
	Url     string
	Headers http.Header
	Level   log.Level

	HandlerCount int
}
