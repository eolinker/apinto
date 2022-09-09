package limiting_stragety

import (
	"github.com/eolinker/apinto/strategy"
	"github.com/eolinker/eosc"
)

var (
	_ eosc.IWorker        = (*Limiting)(nil)
	_ eosc.IWorkerDestroy = (*Limiting)(nil)
)

type Limiting struct {
	id     string
	name   string
	filter strategy.IFilter
}

func (l *Limiting) Destroy() error {
	controller.Del(l.id)
	return nil
}

func (l *Limiting) Id() string {
	return l.id
}

func (l *Limiting) Start() error {
	actuatorSet.Set(l.id, l)
	return nil
}

func (l *Limiting) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	return nil
}

func (l *Limiting) Stop() error {
	actuatorSet.Del(l.id)
	return nil
}

func (l *Limiting) CheckSkill(skill string) bool {
	return false
}
