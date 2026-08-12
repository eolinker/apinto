package pricing_policy

import (
	"github.com/eolinker/apinto/drivers"
	price_calcular "github.com/eolinker/apinto/price-calcular"
	"github.com/eolinker/eosc"
)

var _ eosc.IWorker = (*executor)(nil)

// IPolicyExecutor 额外定义了定价策略驱动的 Worker 外部可访问接口。
type IPolicyExecutor interface {
	// Calculator 返回当前由定价策略实例预编译的高性能计算器。
	Calculator() *price_calcular.Calculator
}

// executor 定价规则 Worker 运行实例。
type executor struct {
	drivers.WorkerBase
	calculator price_calcular.ICalculator
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
	pc, err := price_calcular.NewCalculator(e.Id(), conf.Currency, conf.ContextVariables, conf.AdvancedRules)
	if err != nil {
		return err
	}
	e.calculator = pc
	price_calcular.SetCalculator(e.Name(), pc)
	return nil
}

func (e *executor) Stop() error {
	
	price_calcular.DelCalculator(e.Name())
	e.calculator = nil
	return nil
}

func (e *executor) CheckSkill(skill string) bool {
	// 返回其支持高级规则评估特长
	return skill == "pricing-policy"
}
