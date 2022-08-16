package template

import (
	"github.com/eolinker/apinto/plugin"
	"github.com/eolinker/apinto/template"

	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
)

type Template struct {
	id   string
	name string

	proxyDatas *ProxyDatas
}

func NewTemplate(id string, name string) *Template {
	t := &Template{
		id:   id,
		name: name,

		proxyDatas: NewProxyDatas(),
	}

	return t
}

func (t *Template) Id() string {
	return t.id
}

func (t *Template) Start() error {
	return nil
}

func (t *Template) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	cf, ok := conf.(*Config)
	if !ok {
		return eosc.ErrorConfigIsNil
	}

	t.proxyDatas.Reset(cf.Plugins)

	return nil
}

func (t *Template) Stop() error {
	t.proxyDatas.Destroy()
	return nil
}

func (t *Template) CheckSkill(skill string) bool {
	return template.CheckSkill(skill)
}

func (t *Template) Create(id string, conf map[string]*plugin.Config) eocontext.IChain {
	return t.proxyDatas.Set(id, conf)
}
