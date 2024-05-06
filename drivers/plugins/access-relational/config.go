package access_relational

import "github.com/eolinker/apinto/utils/response"

type Rule struct {
	A string `yaml:"a" json:"a,omitempty" label:"Key A" description:"A key 规则,支持#{}的metrics语法" require:"true"` // ser: #{label service} =>  service: uuid
	B string `yaml:"b" json:"b,omitempty" label:"key B" description:"B Key 规则,支持#{}的metrics语法" require:"true"` // {appid}
}
type Config struct {
	Rules    []*Rule            `yaml:"rules" json:"rules" label:"规则" description:"规则列表, 规则为空时,不执行拦截, 多个规则时,有任意规则通过则均放行, "`
	Response *response.Response `yaml:"response" json:"response" label:"响应内容" description:"请求被拦截时响应的内容"`
}
