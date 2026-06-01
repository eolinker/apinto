package pricing

import (
	"github.com/eolinker/eosc"
)

// IPriceCalculator 价格计算器接口，由 drivers/pricing 实现并注册到全局表。
// 查找键为 provider/resource/phase 三元组，phase 为空表示同步计费。
type IPriceCalculator interface {
	Provider() string
	Resource() string
	Phase() PricingPhase
	Pricing() *ResourcePricing
	RequestFields() map[string]string
	ResponseFields() map[string]string
	Calculate(usage *PricingUsage) (*CalculateResult, error)
}

// priceCalculatorManager 全局计算器注册表。
type priceCalculatorManager struct {
	calculators eosc.Untyped[string, IPriceCalculator]
}

var calculatorManager = newPriceCalculatorManager()

func newPriceCalculatorManager() *priceCalculatorManager {
	return &priceCalculatorManager{
		calculators: eosc.BuildUntyped[string, IPriceCalculator](),
	}
}

// calculatorKey 生成计算器缓存键。同步模式 (phase 为空) 不带 phase 段，
// 异步模式带 phase 段，避免 submit/query 阶段共用同一计算器。
func calculatorKey(provider, resource, phase string) string {
	if phase == "" {
		return provider + "/" + resource
	}
	return provider + "/" + resource + "/" + phase
}

func (m *priceCalculatorManager) set(provider, resource, phase string, calculator IPriceCalculator) {
	m.calculators.Set(calculatorKey(provider, resource, phase), calculator)
}

func (m *priceCalculatorManager) get(provider, resource, phase string) (IPriceCalculator, bool) {
	return m.calculators.Get(calculatorKey(provider, resource, phase))
}

func (m *priceCalculatorManager) del(provider, resource, phase string) {
	m.calculators.Del(calculatorKey(provider, resource, phase))
}

func (m *priceCalculatorManager) list() []IPriceCalculator {
	list := m.calculators.List()
	result := make([]IPriceCalculator, 0, len(list))
	for _, c := range list {
		result = append(result, c)
	}
	return result
}

// SetCalculator 注册价格计算器（drivers/pricing 在 Reset 时调用）。
func SetCalculator(provider, resource, phase string, calculator IPriceCalculator) {
	calculatorManager.set(provider, resource, phase, calculator)
}

// GetCalculator 查找价格计算器（drivers/plugins/billing 在拦截时调用）。
func GetCalculator(provider, resource, phase string) (IPriceCalculator, bool) {
	return calculatorManager.get(provider, resource, phase)
}

// DelCalculator 注销价格计算器（drivers/pricing 在 Stop 时调用）。
func DelCalculator(provider, resource, phase string) {
	calculatorManager.del(provider, resource, phase)
}

// CalculatorList 列出所有已注册计算器（用于诊断和监控）。
func CalculatorList() []IPriceCalculator {
	return calculatorManager.list()
}
