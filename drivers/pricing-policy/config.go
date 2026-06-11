package pricing_policy

import (
	"encoding/json"
	"fmt"

	"github.com/Knetic/govaluate"
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

type Config struct {
	Currency         string               `json:"currency" label:"货币类型" enum:"USD,CNY",default:"USD"`
	ContextVariables map[string]*Variable `json:"context_variables" label:"上下文变量"`
	AdvancedRules    []*Rule              `json:"advanced_rules" label:"高级规则,有顺序"`
}

type Variable struct {
	Source string `json:"source" label:"来源" enum:"request_body,response_body,response_status" default:"response_body"`
	Type   string `json:"type" label:"类型" enum:"string,integer,float,boolean" default:"integer"`
	Path   string `json:"path" label:"Json Path"`
}

type Rule struct {
	ID                 string            `json:"id" label:"规则ID"`
	Name               string            `json:"name" label:"规则名称"`
	Conditions         *Condition        `json:"conditions" label:"条件"`
	Price              map[string]string `json:"price" label:"价格"`
	CostExpression     string            `json:"cost_expression" label:"进货价计算表达式"`
	SaleExpression     string            `json:"sale_expression" label:"销售价计算表达式"`
	OfficialExpression string            `json:"official_expression" label:"官方价计算表达式"`
}

type Condition struct {
	AllOf []*AllOf `json:"all_of" label:"满足全部条件，只能二选一"`
	AnyOf []*AnyOf `json:"any_of" label:"满足任意条件，只能二选一"`
}

type AllOf struct {
	*BasicRule
	AnyOf []*AnyOf `json:"any_of,omitempty"`
}

type AnyOf struct {
	*BasicRule
	AllOfRaw []json.RawMessage `json:"all_of,omitempty"`
	AllOf    []*AllOf          `json:"-"`
}

func (a *AnyOf) UnmarshalJSON(data []byte) error {
	type Alias AnyOf
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(a),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	if len(a.AllOfRaw) > 0 {
		a.AllOf = make([]*AllOf, len(a.AllOfRaw))
		for i, raw := range a.AllOfRaw {
			var allOf AllOf
			if err := json.Unmarshal(raw, &allOf); err != nil {
				return err
			}
			a.AllOf[i] = &allOf
		}
	}
	return nil
}

func (a *AnyOf) MarshalJSON() ([]byte, error) {
	type Alias AnyOf
	if len(a.AllOf) > 0 {
		a.AllOfRaw = make([]json.RawMessage, len(a.AllOf))
		for i, allOf := range a.AllOf {
			raw, err := json.Marshal(allOf)
			if err != nil {
				return nil, err
			}
			a.AllOfRaw[i] = raw
		}
	}
	return json.Marshal((*Alias)(a))
}

type BasicRule struct {
	Key   string `json:"key" label:"变量key"`
	Op    string `json:"op" label:"运算符" enum:">,>=,==,<,<=,!=,in"`
	Value string `json:"value" label:"值"`
	Type  string `json:"type" label:"类型" enum:"string,integer,float,boolean"`
}

// PricingData 描述了从 Redis 缓存中获取的动态资源定价内容的数据结构
type PricingData struct {
	BasicInfo *BasicInfo            `json:"basic_info"`
	Strategy  map[string]*PricePlan `json:"strategy"`
}

type BasicInfo struct {
	Version         string `json:"version"`
	Rely            string `json:"rely"`
	ResourceGroupID string `json:"resource_group_id"`
	TenantID        string `json:"tenant_id"`
}

type PricePlan struct {
	Cost     map[string]float64 `json:"cost"`
	Sale     map[string]float64 `json:"sale"`
	Official map[string]float64 `json:"official"`
}

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

	// 1. 验证上下文变量配置
	for name, variable := range conf.ContextVariables {
		if name == "" {
			return nil, fmt.Errorf("variable name cannot be empty")
		}
		switch variable.Source {
		case "request_body", "response_body":
			if variable.Path == "" {
				return nil, fmt.Errorf("path cannot be empty for variable %s with source %s", name, variable.Source)
			}
		case "response_status":
			// 状态码提取器不需要配置 json path
		default:
			return nil, fmt.Errorf("unsupported source %s for variable %s", variable.Source, name)
		}

		switch variable.Type {
		case "string", "integer", "float", "boolean":
			// 合法的目标类型
		default:
			return nil, fmt.Errorf("unsupported type %s for variable %s", variable.Type, name)
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

		// 提前校验进货价表达式语法
		processedCost := preProcessExpression(rule.CostExpression)
		if _, err := govaluate.NewEvaluableExpression(processedCost); err != nil {
			return nil, fmt.Errorf("invalid cost_expression '%s' in rule %s: %w", rule.CostExpression, rule.ID, err)
		}

		// 提前校验销售价表达式语法
		processedSale := preProcessExpression(rule.SaleExpression)
		if _, err := govaluate.NewEvaluableExpression(processedSale); err != nil {
			return nil, fmt.Errorf("invalid sale_expression '%s' in rule %s: %w", rule.SaleExpression, rule.ID, err)
		}

		// 提前校验官方参考售价表达式语法
		if rule.OfficialExpression != "" {
			processedOfficial := preProcessExpression(rule.OfficialExpression)
			if _, err := govaluate.NewEvaluableExpression(processedOfficial); err != nil {
				return nil, fmt.Errorf("invalid official_expression '%s' in rule %s: %w", rule.OfficialExpression, rule.ID, err)
			}
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
