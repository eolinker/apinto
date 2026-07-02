package pricing_policy_base

import (
	"fmt"
	price_calcular "github.com/eolinker/apinto/price-calcular"
	"github.com/eolinker/eosc"
)

type Config struct {
	Currency             string                   `json:"currency"`
	ContextVariables     price_calcular.Variables `json:"context_variables"`
	AdvancedRules        []*price_calcular.Rule   `json:"advanced_rules"`
	DefaultCalculatorKey string                   `json:"default_calculator_key"`
	CacheKey             string                   `json:"cache_key"`
}

func check(cfg *Config, workers map[eosc.RequireId]eosc.IWorker) error {
	if cfg.Currency == "" { // 默认美元
		cfg.Currency = "USD"
	}

	if cfg.DefaultCalculatorKey != "" {
		return nil
	}
	if cfg.ContextVariables == nil {
		return fmt.Errorf("context_variables cannot be empty")
	}

	err := cfg.ContextVariables.Check()
	if err != nil {
		return err
	}
	if len(cfg.AdvancedRules) < 1 {
		return fmt.Errorf("advanced_rules cannot be empty")
	}
	for _, rule := range cfg.AdvancedRules {
		if rule.SaleExpression == "" {
			return fmt.Errorf("sale_expression cannot be empty")
		}

		if rule.CostExpression == "" {
			return fmt.Errorf("cost_expression cannot be empty")
		}
		// 提前校验销售价表达式语法
		_, err = price_calcular.NewExpression(rule.SaleExpression)
		if err != nil {
			return fmt.Errorf("invalid sale_expression '%s': %w", rule.SaleExpression, err)
		}
		// 提前校验进货价表达式语法
		_, err = price_calcular.NewExpression(rule.CostExpression)
		if err != nil {
			return fmt.Errorf("invalid cost_expression '%s': %w", rule.CostExpression, err)
		}
	}

	return nil
}
