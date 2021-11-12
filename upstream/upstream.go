package upstream

import (
	"time"

	http_service "github.com/eolinker/eosc/http-service"
	"github.com/eolinker/goku/plugin"
)

//CheckSkill 检测目标技能是否符合
func CheckSkill(skill string) bool {
	return skill == "github.com/eolinker/goku/upstream.upstream.IUpstreamCreate"
}

//IUpstream 实现了负载发送请求方法
type IUpstream interface {
	Send(ctx http_service.IHttpContext, retry int, timeout time.Duration) error
}

type IUpstreamCreate interface {
	Create(id string, configs map[string]*plugin.Config, retry int, time time.Duration) (http_service.IChain, error)
}
