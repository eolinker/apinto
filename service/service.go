package service

import (
	"github.com/eolinker/eosc"
	eoscContext "github.com/eolinker/eosc/eocontext"
)

const (
	ServiceSkill = "github.com/eolinker/apinto/service.service.IService"
)

type IService interface {
	eosc.IWorker
	eoscContext.EoApp
	eoscContext.BalanceHandler
	eoscContext.UpstreamHostHandler
	Title() string
}

// CheckSkill 检查目标技能是否符合
func CheckSkill(skill string) bool {
	return skill == ServiceSkill
}
