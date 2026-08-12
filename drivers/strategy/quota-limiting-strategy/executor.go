package quota_limiting_strategy

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

var _ eosc.IWorker = (*executor)(nil)

type executor struct {
	drivers.WorkerBase
}

func (e *executor) Start() error {
	return nil
}

func (e *executor) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	cfg, ok := conf.(*Config)
	if !ok {
		return eosc.ErrorConfigType
	}
	addStrategy(e.Id(), cfg)
	return nil
}

func (e *executor) Stop() error {
	removeStrategy(e.Id())
	return nil
}

func (e *executor) CheckSkill(skill string) bool {
	return false
}
