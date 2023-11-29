package template

import (
	"github.com/eolinker/eosc"
	eoscContext "github.com/eolinker/eosc/eocontext"

	"github.com/eolinker/apinto/plugin"
)

const (
	TemplateSkill = "github.com/eolinker/apinto/template.template.ITemplate"
)

type ITemplate interface {
	eosc.IWorker
	Create(id string, conf map[string]*plugin.Config) eoscContext.IChainPro
}

func CheckSkill(skill string) bool {
	return skill == TemplateSkill
}
