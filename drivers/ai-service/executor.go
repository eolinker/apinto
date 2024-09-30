package ai_service

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/service"

	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
)

var _ service.IService = &executor{}

type executor struct {
	drivers.WorkerBase
	title string
	eocontext.BalanceHandler
}

func (e *executor) PassHost() (eocontext.PassHostMod, string) {
	return eocontext.NodeHost, ""
}

func (e *executor) Title() string {
	return e.title
}

func (e *executor) Start() error {
	return nil
}

func (e *executor) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	//TODO implement me
	panic("implement me")
}

func (e *executor) Stop() error {
	return nil
}

func (e *executor) CheckSkill(skill string) bool {
	return service.CheckSkill(skill)
}
