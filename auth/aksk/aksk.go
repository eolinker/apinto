package aksk

import "github.com/eolinker/eosc"

type aksk struct {
	id     string
	name   string
	labels map[string]string
}

func (a *aksk) Id() string {
	return a.id
}

func (a *aksk) Start() error {
	panic("implement me")
}

func (a *aksk) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
	panic("implement me")
}

func (a *aksk) Stop() error {
	panic("implement me")
}

func (a *aksk) CheckSkill(skill string) bool {
	panic("implement me")
}
