package pricing_policy

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/drivers/pricing-policy/manager"
	"github.com/eolinker/eosc"
)

var _ eosc.IWorker = (*executor)(nil)
var _ manager.IPolicyExecutor = (*executor)(nil)

// IPolicyExecutor 额外定义了定价策略驱动的 Worker 外部可访问接口。
type IPolicyExecutor interface {
	// Calculator 返回当前由定价策略实例预编译的高性能计算器。
	Calculator() *Calculator
}

// executor 定价规则 Worker 运行实例。
type executor struct {
	drivers.WorkerBase
	calculator *Calculator
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
	pc, err := NewCalculator(conf)
	if err != nil {
		return err
	}
	e.calculator = pc
	if policyManager != nil {
		policyManager.Set(e.Id(), e)
	}
	return nil
}

func (e *executor) Stop() error {
	if policyManager != nil {
		policyManager.Del(e.Id())
	}
	e.calculator = nil
	return nil
}

func (e *executor) CheckSkill(skill string) bool {
	// 返回其支持高级规则评估特长
	return skill == "pricing-policy"
}

func (e *executor) Calculator() interface{} {
	return e.calculator
}
