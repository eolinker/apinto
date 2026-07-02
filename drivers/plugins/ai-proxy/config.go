package ai_proxy

import (
	"fmt"
	ai_convert "github.com/eolinker/apinto/ai-convert"
	price_calcular "github.com/eolinker/apinto/price-calcular"
	"github.com/eolinker/eosc"
	"regexp"
)

type Config struct {
	ModelType        string                   `json:"model_type"`
	Labels           map[string]string        `json:"labels"`
	ModelIdFrom      string                   `json:"model_id_from"` // "path" | "body"
	ModelIdKey       string                   `json:"model_id_key"`  // json key or path regex
	Tag              string                   `json:"tag"`
	ContextVariables price_calcular.Variables `json:"context_variables"`
	CostExpression   string                   `json:"cost_expression"`
	SaleExpression   string                   `json:"sale_expression"`
}

func check(cfg *Config, workers map[eosc.RequireId]eosc.IWorker) error {
	if cfg.ModelType == "" {
		cfg.ModelType = string(ai_convert.ModelTypeOpenAIChat)
	}

	if !ai_convert.ModelTypeIsVaild(ai_convert.ModelType(cfg.ModelType)) {
		return fmt.Errorf("model_type %s is not valid", cfg.ModelType)
	}

	if cfg.ModelIdFrom == "" {
		cfg.ModelIdFrom = "body"
	}

	if cfg.ModelIdFrom != "body" && cfg.ModelIdFrom != "path" {
		return fmt.Errorf("model_id_from %s is not valid, must be 'body' or 'path'", cfg.ModelIdFrom)
	}

	if cfg.ModelIdFrom == "body" && cfg.ModelIdKey == "" {
		cfg.ModelIdKey = "$.model"
	}

	if cfg.ModelIdFrom == "path" && cfg.ModelIdKey != "" {
		_, err := regexp.Compile(cfg.ModelIdKey)
		if err != nil {
			return fmt.Errorf("invalid path regex %s: %v", cfg.ModelIdKey, err)
		}
	}
	if cfg.ContextVariables == nil {
		return fmt.Errorf("context_variables cannot be empty")
	}

	err := cfg.ContextVariables.Check()
	if err != nil {
		return err
	}
	if cfg.SaleExpression == "" {
		return fmt.Errorf("sale_expression cannot be empty")
	}

	if cfg.CostExpression == "" {
		return fmt.Errorf("cost_expression cannot be empty")
	}
	// 提前校验销售价表达式语法
	_, err = price_calcular.NewExpression(cfg.SaleExpression)
	if err != nil {
		return fmt.Errorf("invalid sale_expression '%s': %w", cfg.SaleExpression, err)
	}
	// 提前校验进货价表达式语法
	_, err = price_calcular.NewExpression(cfg.CostExpression)
	if err != nil {
		return fmt.Errorf("invalid cost_expression '%s': %w", cfg.CostExpression, err)
	}

	return nil
}
