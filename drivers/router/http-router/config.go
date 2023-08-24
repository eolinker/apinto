package http_router

import (
	"github.com/eolinker/apinto/plugin"
	"github.com/eolinker/eosc"
)

type Config struct {
	Listen  int            `json:"listen" yaml:"listen" title:"port" description:"使用端口" default:"80" label:"端口号" maximum:"65535"`
	Method  []string       `json:"method" yaml:"method" enum:"GET,POST,PUT,DELETE,PATCH,HEAD,OPTIONS" label:"请求方式"`
	Host    []string       `json:"host" yaml:"host" label:"域名"`
	Path    string         `json:"location"`
	Rules   []Rule         `json:"rules" yaml:"rules" label:"路由规则"`
	Service eosc.RequireId `json:"service" yaml:"service" skill:"github.com/eolinker/apinto/service.service.IService" required:"false" empty_label:"使用匿名服务" label:"目标服务"`

	Status int               `json:"status" yaml:"status" label:"响应状态码" switch:"service===''" default:"200" maximum:"1000" minimum:"100"`
	Header map[string]string `json:"header" yaml:"header" label:"响应头部" switch:"service===''"`
	Body   string            `json:"body" yaml:"status" format:"text" label:"响应Body" switch:"service===''"`

	Template  eosc.RequireId `json:"template" yaml:"template" skill:"github.com/eolinker/apinto/template.template.ITemplate" required:"false" label:"插件模版"`
	Websocket bool           `json:"websocket" yaml:"websocket" label:"Websocket" switch:"service!==''"`
	Disable   bool           `json:"disable" yaml:"disable" label:"禁用路由"`
	Plugins   plugin.Plugins `json:"plugins" yaml:"plugins" label:"插件配置"`

	Retry   int               `json:"retry" label:"重试次数" yaml:"retry" switch:"service!==''"`
	TimeOut int               `json:"time_out" label:"超时时间" switch:"service!==''"`
	Labels  map[string]string `json:"labels" label:"路由标签"`
}

// Rule 规则
type Rule struct {
	Type  string `json:"type" yaml:"type" label:"类型" enum:"header,query,cookie"`
	Name  string `json:"name" yaml:"name" label:"参数名"`
	Value string `json:"value" yaml:"value" label:"值规" `
}
