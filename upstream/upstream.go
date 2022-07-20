package upstream

import (
	"github.com/eolinker/eosc/context"
	"time"
)

//CheckSkill 检测目标技能是否符合
func CheckSkill(skill string) bool {
	return skill == "github.com/eolinker/apinto/upstream.upstream.IUpstream"
}

type IUpstreamHandler interface {
	context.IChain
}

type IUpstream interface {
	Create(id string, retry int, time time.Duration) (IUpstreamHandler, error)
}
