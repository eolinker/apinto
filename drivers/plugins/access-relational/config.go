package access_relational

import "github.com/eolinker/apinto/utils/response"

type Rule struct {
	KeyRule    string `yaml:"key_rule" json:"keyRule,omitempty" label:"namespace" description:"namespace, 支持${}, 对应hash的key" require:"true"` // ser: ${label service} =>  service: uuid
	AccessRule string `yaml:"access" json:"fieldRule,omitempty" label:"授权key" description:"放行值规则,支持${}, 对应 hash的field" require:"true"`       // {appid}
}
type Config struct {
	Rules    []*Rule            `yaml:"rules" json:"rules" label:"规则" description:"规则列表, 规则为空时,不执行拦截, 多个规则时,有任意规则通过,均放行"`
	Response *response.Response `yaml:"response" json:"response" label:"响应内容" description:"请求被拦截时响应的内容"`
}
