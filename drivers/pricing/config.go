package pricing_driver

import (
	"fmt"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/pricing"
	"github.com/eolinker/eosc"
)

// Config 价格计算器配置。
//
// 字段含义：
//   - Provider:        供应商或服务分组（AI 场景=厂商，API 场景=服务名）
//   - Resource:        资源（AI 场景=模型名，API 场景=接口/端点标识）
//   - Phase:           计费阶段，同步留空，异步任务用 submit/query
//   - Pricing:         资源定价配置
//   - DynamicStrategy: 可选的动态调价策略，应用于基础与阶梯单价
//   - RequestFields:   计算器级请求字段提取规则（高优先级覆盖插件级）
//   - ResponseFields:  计算器级响应字段提取规则（高优先级覆盖插件级）
type Config struct {
	Provider        string                  `json:"provider" label:"供应商" description:"供应商或服务分组（AI=厂商，API=服务名）"`
	Resource        string                  `json:"resource" label:"资源" description:"AI=模型名；API=接口/端点标识"`
	Phase           pricing.PricingPhase    `json:"phase" label:"计费阶段" description:"空(同步)、submit(提交预扣)、query(查询实扣)" enum:",submit,query"`
	Pricing         pricing.ResourcePricing `json:"pricing" label:"资源定价" description:"该资源的定价配置"`
	DynamicStrategy *DynamicPricingStrategy `json:"dynamic_strategy,omitempty" label:"动态调价策略" description:"可选的动态调价策略"`
	RequestFields   map[string]string       `json:"request_fields,omitempty" label:"请求字段提取" description:"从请求体中提取字段的JSONPath映射"`
	ResponseFields  map[string]string       `json:"response_fields,omitempty" label:"响应字段提取" description:"从响应体中提取字段的JSONPath映射"`
}

// DynamicPricingStrategy 动态调价策略。
//   - ReferencePrice=official：Discount 必须 < 1.0（在官方价基础上打折）
//   - ReferencePrice=purchase：Discount 必须 > 1.0（在采购价基础上加价）
type DynamicPricingStrategy struct {
	Name           string             `json:"name" label:"策略名称"`
	ReferencePrice pricing.PriceLevel `json:"reference_price" label:"参考价格" description:"参考价格类型" enum:"official,purchase"`
	Discount       float64            `json:"discount" label:"折扣比例"`
}

func checkConfig(v interface{}) (*Config, error) {
	cfg, ok := v.(*Config)
	if !ok {
		return nil, eosc.ErrorConfigType
	}

	if cfg.Provider == "" {
		return nil, fmt.Errorf("provider is required")
	}
	if cfg.Resource == "" {
		return nil, fmt.Errorf("resource is required")
	}
	if cfg.Pricing.Mode == "" {
		return nil, fmt.Errorf("pricing.mode is required")
	}
	// resource 字段优先以 Config.Resource 为准；若 Pricing.Resource 留空则填充。
	if cfg.Pricing.Resource == "" {
		cfg.Pricing.Resource = cfg.Resource
	}

	if cfg.DynamicStrategy != nil {
		switch cfg.DynamicStrategy.ReferencePrice {
		case pricing.PriceLevelPurchase:
			if cfg.DynamicStrategy.Discount <= 1.0 {
				return nil, fmt.Errorf("dynamic strategy discount must be > 1.0 when reference price is purchase")
			}
		case pricing.PriceLevelOfficial:
			if cfg.DynamicStrategy.Discount >= 1.0 {
				return nil, fmt.Errorf("dynamic strategy discount must be < 1.0 when reference price is official")
			}
		}
	}

	return cfg, nil
}

// applyDiscount 在基础和阶梯单价上整体应用折扣系数。
// 函数返回新的 ResourcePricing 副本，避免修改原配置。
func applyDiscount(p pricing.ResourcePricing, strategy *DynamicPricingStrategy) pricing.ResourcePricing {
	d := strategy.Discount
	p.Input *= d
	p.Output *= d
	p.CachedInput *= d
	p.PerCall *= d
	p.PerSecond *= d
	for i := range p.AdvancedRules {
		rule := &p.AdvancedRules[i]
		rule.Pricing.Input *= d
		rule.Pricing.Output *= d
		rule.Pricing.CachedInput *= d
		rule.Pricing.PerCall *= d
		rule.Pricing.PerSecond *= d
	}
	return p
}

// Create 实例化 worker。
func Create(id, name string, v *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	cfg, err := checkConfig(v)
	if err != nil {
		return nil, err
	}
	w := &executor{
		WorkerBase: drivers.Worker(id, name),
	}
	if err := w.reset(cfg); err != nil {
		return nil, err
	}
	return w, nil
}
