package service

import (
	"github.com/eolinker/apinto/plugin"
	"github.com/eolinker/eosc"
	eoscContext "github.com/eolinker/eosc/eocontext"
)

type IService interface {
	eosc.IWorker
	eoscContext.EoApp
	eoscContext.BalanceHandler
}

type ITemplate interface {
	eosc.IWorker
	Create(id string, conf map[string]*plugin.Config) eoscContext.IChain
}

//CheckSkill 检查目标技能是否符合
func CheckSkill(skill string) bool {
	return skill == "github.com/eolinker/apinto/service.service.IService"
}
