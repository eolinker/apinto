package service

import (
	"github.com/eolinker/eosc"
	eoscContext "github.com/eolinker/eosc/eocontext"
	"github.com/eolinker/eosc/utils/config"
)

var (
	ServiceSkill string
)

func init() {
	var s IService
	ServiceSkill = config.TypeNameOf(&s)

}

type IService interface {
	eosc.IWorker
	eoscContext.EoApp
	eoscContext.BalanceHandler
}

//CheckSkill 检查目标技能是否符合
func CheckSkill(skill string) bool {
	return skill == ServiceSkill
}
