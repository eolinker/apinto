package service_http

import (
	"github.com/eolinker/eosc"
)

//Config service_http驱动配置
type Config struct {
	id         string
	Name       string `json:"name"`
	Driver     string `json:"driver"`
	Desc       string `json:"desc"`
	Timeout    int64  `json:"timeout"`
	Retry      int    `json:"retry"`
	Scheme     string `json:"scheme"`
	RewriteURL string `json:"rewrite_url"`
	//Auth       []eosc.RequireId `json:"auth" skill:"github.com/eolinker/goku-eosc/auth.auth.IAuth"`
	Upstream eosc.RequireId `json:"upstream" skill:"github.com/eolinker/goku-eosc/upstream.upstream.IUpstream"`
}
