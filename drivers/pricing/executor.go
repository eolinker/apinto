package pricing_driver

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/pricing"
	"github.com/eolinker/eosc"
)

var _ eosc.IWorker = (*executor)(nil)

// executor 价格计算器 worker，承担定价配置的生命周期管理。
//
// 生命周期：
//   - Reset：根据配置构建计算器并注册到顶层 pricing 全局表（覆盖旧实例）
//   - Stop：从全局表注销
//   - Destroy：清空内部引用，便于 GC
type executor struct {
	drivers.WorkerBase
	provider   string
	resource   string
	phase      pricing.PricingPhase
	calculator pricing.IPriceCalculator
}

// Start 启动 worker。计算器在 Reset 时已注册，Start 阶段无需额外动作。
func (e *executor) Start() error {
	return nil
}

// Reset 处理配置变更：销毁旧注册项后重新构建并注册新的计算器。
func (e *executor) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	cfg, err := checkConfig(conf)
	if err != nil {
		return err
	}
	// 先注销旧三元组（provider/resource/phase 可能与新配置不同）
	if e.calculator != nil {
		pricing.DelCalculator(e.provider, e.resource, string(e.phase))
	}
	return e.reset(cfg)
}

// reset 根据 Config 构建计算器并注册到全局表。
func (e *executor) reset(conf *Config) error {
	resourcePricing := conf.Pricing
	if conf.DynamicStrategy != nil {
		resourcePricing = applyDiscount(resourcePricing, conf.DynamicStrategy)
	}

	pc := NewPriceCalculator(conf.Provider, resourcePricing, conf.Phase, conf.RequestFields, conf.ResponseFields)
	e.calculator = pc
	e.provider = conf.Provider
	e.resource = conf.Resource
	e.phase = conf.Phase
	pricing.SetCalculator(e.provider, e.resource, string(e.phase), e.calculator)
	return nil
}

// Stop 注销当前计算器。允许重复调用：DelCalculator 内部对不存在的键安全。
func (e *executor) Stop() error {
	pricing.DelCalculator(e.provider, e.resource, string(e.phase))
	e.calculator = nil
	return nil
}

// Destroy 释放内部引用。
func (e *executor) Destroy() {
	e.calculator = nil
}

// CheckSkill 校验 skill 是否为价格计算器技能，用于其他 worker 引用本 worker。
func (e *executor) CheckSkill(skill string) bool {
	return pricing.CheckCalculatorSkill(skill)
}
