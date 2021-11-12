package upstream

import (
	http_context "github.com/eolinker/goku/node/http-context"
	"github.com/eolinker/goku/plugin"
	"github.com/valyala/fasthttp"
)

//CheckSkill 检测目标技能是否符合
func CheckSkill(skill string) bool {
	return skill == "github.com/eolinker/goku/upstream.upstream.IUpstreamCreate"
}

//IUpstream 实现了负载发送请求方法
type IUpstream interface {
	Send(ctx *http_context.Context) (*fasthttp.Response, error)
}

type IUpstreamCreate interface {
	Create(id string, configs map[string]*plugin.Config) IUpstream
}
