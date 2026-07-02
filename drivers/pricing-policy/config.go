package pricing_policy

import (
	"fmt"
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/price-calcular"
	"github.com/eolinker/eosc"
)

type Config struct {
	Currency         string                   `json:"currency" label:"货币类型" enum:"USD,CNY",default:"USD"`
	ContextVariables price_calcular.Variables `json:"context_variables" label:"上下文变量"`
	AdvancedRules    []*price_calcular.Rule   `json:"advanced_rules" label:"高级规则,有顺序"`
}

// PricingData 描述了从 Redis 缓存中获取的动态资源定价内容的数据结构

func checkConfig(v interface{}) (*Config, error) {
	conf, ok := v.(*Config)
	if !ok {
		return nil, eosc.ErrorConfigType
	}

	if conf.Currency == "" {
		conf.Currency = "USD"
	} else if conf.Currency != "USD" && conf.Currency != "CNY" {
		return nil, fmt.Errorf("unsupported currency: %s", conf.Currency)
	}

	if conf.ContextVariables != nil {
		err := conf.ContextVariables.Check()
		if err != nil {
			return nil, err
		}
	}

	// 2. 验证高级计费规则及公式语法
	for i, rule := range conf.AdvancedRules {
		if rule.ID == "" {
			return nil, fmt.Errorf("rule id cannot be empty at index %d", i)
		}
		if rule.CostExpression == "" {
			return nil, fmt.Errorf("cost_expression cannot be empty in rule %s", rule.ID)
		}
		if rule.SaleExpression == "" {
			return nil, fmt.Errorf("sale_expression cannot be empty in rule %s", rule.ID)
		}
		if rule.OfficialExpression == "" {
			rule.OfficialExpression = rule.CostExpression
		}

		// 提前校验进货价表达式语法
		_, err := price_calcular.NewExpression(rule.CostExpression)
		if err != nil {
			return nil, fmt.Errorf("invalid cost_expression '%s' in rule %s: %w", rule.CostExpression, rule.ID, err)
		}

		// 提前校验销售价表达式语法
		_, err = price_calcular.NewExpression(rule.SaleExpression)
		if err != nil {
			return nil, fmt.Errorf("invalid sale_expression '%s' in rule %s: %w", rule.SaleExpression, rule.ID, err)
		}

		// 提前校验官方参考售价表达式语法
		_, err = price_calcular.NewExpression(rule.OfficialExpression)
		if err != nil {
			return nil, fmt.Errorf("invalid official_expression '%s' in rule %s: %w", rule.OfficialExpression, rule.ID, err)
		}
	}

	return conf, nil
}
func Create(id, name string, v *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	cfg, err := checkConfig(v)
	if err != nil {
		return nil, err
	}
	w := &executor{
		WorkerBase: drivers.Worker(id, name),
	}
	err = w.reset(cfg)
	return w, err
}
