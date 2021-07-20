package service_http

import (
	"github.com/eolinker/eosc"
)

type Config struct {
	id         string
	Name       string         `json:"name"`
	Driver     string         `json:"driver"`
	Desc       string         `json:"desc"`
	Timeout    int64          `json:"timeout"`
	Retry      int            `json:"retry"`
	Scheme     string         `json:"scheme"`
	RewriteUrl string         `json:"rewrite_url"`
	Upstream   eosc.RequireId `json:"upstream" skill:"github.com/eolinker/goku-eosc/upstream.upstream.IUpstream"`
}
