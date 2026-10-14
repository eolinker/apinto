package credit

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

var _ eosc.IWorker = (*executor)(nil)

// ICreditWorker 代金券规则 Worker 暴露的接口
type ICreditWorker interface {
	Rule() *CompiledRule
}

type executor struct {
	drivers.WorkerBase
	rule *CompiledRule
}

func (e *executor) Start() error {
	return nil
}

func (e *executor) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	cfg, err := checkConfig(conf)
	if err != nil {
		return err
	}
	_ = e.Stop()
	return e.reset(cfg)
}

func (e *executor) reset(conf *Config) error {
	rule := CompileRule(e.Name(), conf)
	e.rule = rule
	SetRule(e.Id(), rule)
	return nil
}

func (e *executor) Stop() error {
	DelRule(e.Id())
	e.rule = nil
	return nil
}

func (e *executor) CheckSkill(skill string) bool {
	return skill == "credit"
}

func (e *executor) Rule() *CompiledRule {
	return e.rule
}
