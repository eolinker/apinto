package variable_relation

import "github.com/eolinker/apinto/utils/response"

// Config 变量关系插件配置
type Config struct {
	Rules    []*Rule            `yaml:"rules" json:"rules" label:"规则" description:"规则列表, 规则为空时,不执行拦截, 多个规则时,有任意规则通过则均放行, "`
	Response *response.Response `yaml:"response" json:"response" label:"响应内容" description:"请求被拦截时响应的内容"`
}

type Rule struct {
	Key        string `yaml:"key" json:"key" label:"Key模板" description:"用于获取自定义变量的Key模板，支持{var}语法" require:"true"`
	InnerKey   string `yaml:"inner_key" json:"inner_key" label:"内部Key" description:"从上下文获取值的Key，支持{var}语法" require:"true"`
	ValueLabel string `yaml:"value_label" json:"value_label" label:"值标签名称" description:"将映射结果存储到上下文的标签名称" `
}
