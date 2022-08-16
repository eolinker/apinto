package template

import (
	"github.com/eolinker/apinto/plugin"
	"github.com/eolinker/eosc"
	eoscContext "github.com/eolinker/eosc/eocontext"
)

const (
	TemplateSkill = "github.com/eolinker/apinto/template.template.ITemplate"
)

type ITemplate interface {
	eosc.IWorker
	Create(id string, conf map[string]*plugin.Config) eoscContext.IChain
}

func CheckSkill(skill string) bool {
	return skill == TemplateSkill
}
