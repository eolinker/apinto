package http_router

import (
	"github.com/eolinker/apinto/plugin"
	"github.com/eolinker/eosc"
)

type Config struct {
	Listen   int                       `json:"listen" yaml:"listen" title:"port" description:"使用端口" default:"80" label:"端口号" maximum:"65535"`
	Method   []string                  `json:"method" yaml:"method" enum:"GET,POST,PUT,DELETE,PATH,HEAD,OPTIONS" label:"请求方式"`
	Host     []string                  `json:"host" yaml:"host" label:"域名"`
	Path     string                    `json:"location"`
	Rules    []Rule                    `json:"rules" yaml:"rules" label:"路由规则"`
	Service  eosc.RequireId            `json:"Service" yaml:"Service" skill:"github.com/eolinker/apinto/service.service.IService" required:"true" label:"目标服务"`
	Template eosc.RequireId            `json:"template" yaml:"template" skill:"github.com/eolinker/apinto/template.template.ITemplate" required:"false" label:"插件模版"`
	Disable  bool                      `json:"disable" yaml:"disable" label:"禁用路由"`
	Plugins  map[string]*plugin.Config `json:"plugins" yaml:"plugins" label:"插件配置"`
	Retry    int                       `json:"retry" label:"重试次数" yaml:"retry"`
	TimeOut  int                       `json:"time_out" label:"超时时间"`
}

//Rule 规则
type Rule struct {
	Type  string `json:"type" yaml:"type" label:"类型" enum:"header,query,cookie"`
	Name  string `json:"name" yaml:"name" label:"参数名"`
	Value string `json:"value" yaml:"value" label:"值规" `
}
