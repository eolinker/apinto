package service_http

import (
	"strings"

	"github.com/eolinker/eosc"
)

//Config service_http驱动配置
type Config struct {
	id          string
	Name        string           `json:"name"`
	Driver      string           `json:"driver"`
	Desc        string           `json:"desc"`
	Timeout     int64            `json:"timeout"`
	Retry       int              `json:"retry"`
	Scheme      string           `json:"scheme"`
	RewriteURL  string           `json:"rewrite_url"`
	ProxyMethod string           `json:"proxy_method"`
	Auth        []eosc.RequireId `json:"auth" skill:"github.com/eolinker/goku/auth.auth.IAuth"`
	Upstream    eosc.RequireId   `json:"upstream" skill:"github.com/eolinker/goku/upstream.upstream.IUpstream" require:"false"`
}

var validMethods = []string{
	"GET",
	"POST",
	"PUT",
	"DELETE",
	"PATCH",
	"HEAD",
	"OPTIONS",
}

var validScheme = []string{
	"HTTP",
	"HTTPS",
}

func (c *Config) rebuild() {
	if c.Retry < 0 {
		c.Retry = 0
	}

	if !checkValidParams(strings.ToUpper(c.ProxyMethod), validMethods) {
		c.ProxyMethod = ""
	}

	if !checkValidParams(strings.ToUpper(c.Scheme), validScheme) {
		c.Scheme = "http"
	}
}

func checkValidParams(data string, params []string) bool {
	for _, p := range params {
		if data == p {
			return true
		}
	}
	return false
}
