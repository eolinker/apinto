package openAI

import (
	"time"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
)

type executor struct {
	drivers.WorkerBase
}

func (e *executor) Select(ctx eocontext.EoContext) (eocontext.INode, int, error) {
	//TODO implement me
	panic("implement me")
}

func (e *executor) Scheme() string {
	//TODO implement me
	panic("implement me")
}

func (e *executor) TimeOut() time.Duration {
	//TODO implement me
	panic("implement me")
}

func (e *executor) Nodes() []eocontext.INode {
	//TODO implement me
	panic("implement me")
}

func (e *executor) Start() error {
	return nil
}

func (e *executor) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	//TODO implement me
	panic("implement me")
}

func (e *executor) Stop() error {
	//TODO implement me
	panic("implement me")
}

func (e *executor) CheckSkill(skill string) bool {
	//TODO implement me
	panic("implement me")
}
