package access_relation_redis

import (
	"github.com/eolinker/apinto/utils/response"
	"github.com/eolinker/eosc"
)

type Rule struct {
	RedisKey string `yaml:"redis_key" json:"redis_key,omitempty" label:"Redis Key模板" description:"Redis 存储 Key 模板，支持 {var} 占位符" require:"true"`
	Label    string `yaml:"label" json:"label,omitempty" label:"标签" description:"标签"`
}

type Config struct {
	Cache    eosc.RequireId     `json:"cache" label:"缓存资源" skill:"github.com/eolinker/apinto/resources.resources.ICache" required:"false" description:"Redis 缓存资源 ID"`
	Rules    []*Rule            `yaml:"rules" json:"rules" label:"规则" description:"规则列表，规则为空时，不执行拦截。多个规则时，有任意规则通过则均放行"`
	Response *response.Response `yaml:"response" json:"response" label:"响应内容" description:"请求被拦截时响应的内容"`
}
