package template

import (
	"github.com/eolinker/apinto/plugin"
	"github.com/eolinker/eosc"
	eoscContext "github.com/eolinker/eosc/eocontext"
	"github.com/eolinker/eosc/utils/config"
)

var (
	TemplateSkill string
)

func init() {
	var t ITemplate
	TemplateSkill = config.TypeNameOf(&t)
}

type ITemplate interface {
	eosc.IWorker
	Create(id string, conf map[string]*plugin.Config) eoscContext.IChain
}

func CheckSkill(skill string) bool {
	return skill == TemplateSkill
}
