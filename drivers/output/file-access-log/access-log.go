package file_access_log

import (
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/formatter"
)

type accessLog struct {
	id        string
	formatter formatter.Config
}

func (a *accessLog) Destroy() {
	panic("implement me")
}

func (a *accessLog) Id() string {
	return a.id
}

func (a *accessLog) Start() error {
	panic("implement me")
}

func (a *accessLog) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
	panic("implement me")
}

func (a *accessLog) Stop() error {
	panic("implement me")
}

func (a *accessLog) CheckSkill(skill string) bool {
	panic("implement me")
}
