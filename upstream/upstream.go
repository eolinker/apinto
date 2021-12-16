package upstream

import (
	"time"

	http_service "github.com/eolinker/eosc/http-service"
	"github.com/eolinker/goku/plugin"
)

//CheckSkill 检测目标技能是否符合
func CheckSkill(skill string) bool {
	return skill == "github.com/eolinker/goku/upstream.upstream.IUpstream"
}

type IUpstreamHandler interface {
	http_service.IChain
}
type IUpstream interface {
	//plugin.IPluginConfig
	Create(id string, configs map[string]*plugin.Config, retry int, time time.Duration) (IUpstreamHandler, error)
}
