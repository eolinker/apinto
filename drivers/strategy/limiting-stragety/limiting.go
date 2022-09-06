package limiting_stragety

import (
	"github.com/eolinker/apinto/strategy"
	"github.com/eolinker/eosc"
)

type Limiting struct {
	id     string
	name   string
	filter strategy.IFilter
}

func (l *Limiting) Id() string {
	return l.id
}

func (l *Limiting) Start() error {
	return nil
}

func (l *Limiting) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	return nil
}

func (l *Limiting) Stop() error {
	return nil
}

func (l *Limiting) CheckSkill(skill string) bool {
	return false
}
