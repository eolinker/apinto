package pricing_policy

import (
	"errors"
	"fmt"
	context_label "github.com/eolinker/apinto/utils/context-label"
	"regexp"
	"strconv"
	"strings"

	"github.com/Knetic/govaluate"
	eoscContext "github.com/eolinker/eosc/eocontext"
)

// exprRegex 正则表达式用于匹配类似 #cost.per_second, #sale.input_token 价格属性前缀。
// 在传入 govaluate 前需要将前缀及点 "#cost." 整体替换为合法的变量命名（如 "cost_"），使其兼容 govaluate 词法。
var exprRegex = regexp.MustCompile(`#([a-zA-Z0-9]+)\.([a-zA-Z0-9_]+)`)

// preProcessExpression 将表达式中的 "#prefix.field" 替换为 "prefix_field" 格式，使其兼容 govaluate 词法。
func preProcessExpression(expr string) string {
	return exprRegex.ReplaceAllString(expr, "${1}_${2}")
}

// ProcessedRule 描述一个经过预编译及性能优化的高级定价规则实例。
type ProcessedRule struct {
	id              string                         // 规则 ID
	name            string                         // 规则名称
	conditions      *Condition                     // 规则的递归触发条件树
	costOrgExpr     string                         // 进货价计算表达式
	saleOrgExpr     string                         // 销售价计算表达式
	officialOrgExpr string                         // 官方价计算表达式
	costExpr        *govaluate.EvaluableExpression // 预编译后的进货价（成本）计算表达式
	saleExpr        *govaluate.EvaluableExpression // 预编译后的销售价（计费）计算表达式
	officialExpr    *govaluate.EvaluableExpression // 预编译后的官方参考价计算表达式
}

// CalculateResult 代表高级规则计算后输出的价格评估结果。
type CalculateResult struct {
	Cost     float64 // 进货（成本）价
	Sale     float64 // 销售（计费）价
	Official float64 // 官方参考建议价
}

// Calculator 高性能定价计算器，其计算所依赖的全部上下文计量属性均通过配置的变量动态提取。
type Calculator struct {
	currency           string              // 扣减所采用的的货币类型（USD, CNY）
	variablesExtractor *VariablesExtractor // 该计算器关联的上下文变量提取调度器
	rules              []*ProcessedRule    // 按顺序配置的 ProcessedRule 定价规则链
}

// NewCalculator 根据传入的基础 Config 数据，初始化并预编译一套高性能的定价计算器。
func NewCalculator(conf *Config) (*Calculator, error) {
	// 初始化自定义变量提取器
	extractor, err := NewVariablesExtractor(conf.ContextVariables)
	if err != nil {
		return nil, fmt.Errorf("create variables extractor error: %w", err)
	}

	processedRules := make([]*ProcessedRule, 0, len(conf.AdvancedRules))
	for _, rule := range conf.AdvancedRules {
		var costExpr, saleExpr, officialExpr *govaluate.EvaluableExpression

		// 预编译进货价计算公式
		if rule.CostExpression != "" {
			processed := preProcessExpression(rule.CostExpression)
			expr, err := govaluate.NewEvaluableExpression(processed)
			if err != nil {
				return nil, fmt.Errorf("compile cost_expression '%s' error: %w", rule.CostExpression, err)
			}
			costExpr = expr
		}

		// 预编译销售价计算公式
		if rule.SaleExpression != "" {
			processed := preProcessExpression(rule.SaleExpression)
			expr, err := govaluate.NewEvaluableExpression(processed)
			if err != nil {
				return nil, fmt.Errorf("compile sale_expression '%s' error: %w", rule.SaleExpression, err)
			}
			saleExpr = expr
		}

		// 预编译官方价计算公式
		if rule.OfficialExpression != "" {
			processed := preProcessExpression(rule.OfficialExpression)
			expr, err := govaluate.NewEvaluableExpression(processed)
			if err != nil {
				return nil, fmt.Errorf("compile official_expression '%s' error: %w", rule.OfficialExpression, err)
			}
			officialExpr = expr
		}

		processedRules = append(processedRules, &ProcessedRule{
			id:              rule.ID,
			name:            rule.Name,
			conditions:      rule.Conditions,
			costOrgExpr:     rule.CostExpression,
			saleOrgExpr:     rule.SaleExpression,
			officialOrgExpr: rule.OfficialExpression,
			costExpr:        costExpr,
			saleExpr:        saleExpr,
			officialExpr:    officialExpr,
		})
	}

	return &Calculator{
		currency:           conf.Currency,
		variablesExtractor: extractor,
		rules:              processedRules,
	}, nil
}

// Currency 获取计算器币种类型。
func (c *Calculator) Currency() string {
	return c.currency
}

// VariablesExtractor 获取关联的变量提取调度器实例。
func (c *Calculator) VariablesExtractor() *VariablesExtractor {
	return c.variablesExtractor
}

// Calculate 根据 EoContext 进行自定义属性抽取，结合从 Redis 获取的最新的资源定价内容执行公式计费。
func (c *Calculator) Calculate(ctx eoscContext.EoContext, enableBalance bool, pricingData *PricingData, extendKeys ...string) (*CalculateResult, error) {

	// 动态抓取当前 EoContext 下的所有已配置变量集
	vars := c.variablesExtractor.ExtractAll(ctx, extendKeys...)
	if len(vars) < 1 {
		return nil, errors.New("no context variables extracted")
	}
	oldVar := context_label.GetPriceVariables(ctx)
	for k, v := range oldVar {
		_, ok := vars[k]
		if ok {
			continue
		}
		vars[k] = v
	}

	context_label.SetPriceVariables(ctx, vars)
	if !enableBalance {
		return &CalculateResult{}, nil
	}

	if pricingData == nil {
		return nil, errors.New("redis pricing data is required but got nil")
	}

	// 1. 按配置的高级计费规则顺序，依次匹配条件
	var matchedRule *ProcessedRule
	for _, rule := range c.rules {
		if matchCondition(rule.conditions, vars) {
			matchedRule = rule
			break
		}
	}

	if matchedRule == nil {
		return nil, errors.New("no advanced rules matched the given context variables")
	}

	context_label.SetChargeRule(ctx, matchedRule.id)
	context_label.SetExprCost(ctx, matchedRule.costOrgExpr)
	context_label.SetExprSale(ctx, matchedRule.saleOrgExpr)
	context_label.SetExprOfficial(ctx, matchedRule.officialOrgExpr)
	context_label.SetPriceMatchRule(ctx, matchedRule)
	context_label.SetPriceMatchCondition(ctx, matchedRule.conditions)
	context_label.SetPriceCurrency(ctx, c.Currency())
	params, err := c.generateParams(ctx, vars, pricingData, matchedRule)
	if err != nil {
		return nil, err
	}
	return c.CalculateFromVariables(params, matchedRule)
}

// CalculateFromChunk 根据流式的单个响应原始字节块进行提取转换，结合从 Redis 获取的最新的资源定价内容执行公式计费。
func (c *Calculator) CalculateFromChunk(ctx eoscContext.EoContext, enableBalance bool, chunk []byte, pricingData *PricingData) (*CalculateResult, error) {

	// 从流式 Chunk 里解包和转换出最新的计量参数变量集
	newVars := c.variablesExtractor.ExtractAllFromChunk(ctx, chunk)
	if len(newVars) < 1 {
		return &CalculateResult{}, nil
	}
	vars := context_label.GetPriceVariables(ctx)
	if vars == nil {
		vars = make(map[string]interface{})
	}
	for k, v := range newVars {
		vars[k] = v
	}
	context_label.SetPriceVariables(ctx, vars)
	if !enableBalance {
		return &CalculateResult{}, nil
	}
	if pricingData == nil {
		return nil, errors.New("redis pricing data is required but got nil")
	}
	var mr *ProcessedRule
	matchedRule := context_label.GetPriceMatchRule(ctx)
	if matchedRule == nil {
		for _, rule := range c.rules {
			if matchCondition(rule.conditions, vars) {
				matchedRule = rule
				break
			}
		}
		if matchedRule == nil {
			return nil, errors.New("no advanced rules matched the given context variables")
		}
		mr = matchedRule.(*ProcessedRule)
		context_label.SetChargeRule(ctx, mr.id)
		context_label.SetExprCost(ctx, mr.costOrgExpr)
		context_label.SetExprSale(ctx, mr.saleOrgExpr)
		context_label.SetExprOfficial(ctx, mr.officialOrgExpr)
		context_label.SetPriceMatchCondition(ctx, mr.conditions)
		context_label.SetPriceMatchRule(ctx, matchedRule)
	} else {
		mr = matchedRule.(*ProcessedRule)
	}

	params, err := c.generateParams(ctx, vars, pricingData, mr)
	if err != nil {
		return nil, err
	}
	return c.CalculateFromVariables(params, mr)
}

// 生成计费参数
func (c *Calculator) generateParams(ctx eoscContext.EoContext, vars map[string]interface{}, pricingData *PricingData, matchedRule *ProcessedRule) (map[string]interface{}, error) {
	var plan *PricePlan
	if matchedRule.id != "" {
		// 从缓存中获取对应的价格包
		plan = pricingData.Strategy[matchedRule.id]
	}
	if plan == nil {
		return nil, fmt.Errorf("neither price plan for rule '%s' nor base pricing plan is configured in strategy", matchedRule.id)
	}
	// 3. 构建公式运行所需的上下文变量 map (合并自定义提取变量与绑定的动态价格变量)
	params := make(map[string]interface{}, len(vars)+len(plan.Cost)+len(plan.Sale)+len(plan.Official))

	for k, v := range vars {
		params[k] = v
	}
	for k, v := range plan.Cost {
		params["cost_"+k] = v
	}
	for k, v := range plan.Sale {
		params["sale_"+k] = v
	}
	for k, v := range plan.Official {
		params["official_"+k] = v
	}
	context_label.SetPriceCost(ctx, plan.Cost)
	context_label.SetPriceSale(ctx, plan.Sale)
	context_label.SetPriceOfficial(ctx, plan.Official)
	return params, nil
}

// CalculateFromVariables 根据给定的参数变量字典与从 Redis 中读取的最新资源定价大 JSON，
// 执行高级规则链条的条件评估匹配，绑定价格并进行表达式最终计费计算。
func (c *Calculator) CalculateFromVariables(params map[string]interface{}, matchedRule *ProcessedRule) (*CalculateResult, error) {
	// 计算成本价
	costPrice, err := evaluateExpression(matchedRule.costExpr, params)
	if err != nil {
		return nil, fmt.Errorf("evaluate cost_expression failed: %w", err)
	}

	// 计算销售价
	salePrice, err := evaluateExpression(matchedRule.saleExpr, params)
	if err != nil {
		return nil, fmt.Errorf("evaluate sale_expression failed: %w", err)
	}

	// 6. 计算官方建议售价
	officialPrice := salePrice
	if matchedRule.officialExpr != nil {
		op, err := evaluateExpression(matchedRule.officialExpr, params)
		if err == nil {
			officialPrice = op
		}
	}

	return &CalculateResult{
		Cost:     costPrice,
		Sale:     salePrice,
		Official: officialPrice,
	}, nil
}

// evaluateExpression 使用 govaluate 执行浮点数公式计算。
func evaluateExpression(expr *govaluate.EvaluableExpression, params map[string]interface{}) (float64, error) {
	if expr == nil {
		return 0, nil
	}
	result, err := expr.Evaluate(params)
	if err != nil {
		return 0, err
	}
	switch val := result.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	default:
		str := fmt.Sprintf("%v", result)
		f, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return 0, fmt.Errorf("expression result is not a number: %v", result)
		}
		return f, nil
	}
}

// matchBasicRule 负责根据指定的逻辑类型，评估一个基础条件（BasicRule）是否在当前的 parameters 下成立。
func matchBasicRule(rule *BasicRule, params map[string]interface{}) bool {
	val, exists := params[rule.Key]
	if !exists {
		return false
	}

	switch rule.Type {
	case "array":
		var actual []interface{}
		switch v := val.(type) {
		case []interface{}:
			actual = v
		case []string:
			actual = make([]interface{}, len(v))
			for i, s := range v {
				actual[i] = s
			}
		case []int64:
			actual = make([]interface{}, len(v))
			for i, s := range v {
				actual[i] = s
			}
		case []float64:
			actual = make([]interface{}, len(v))
			for i, s := range v {
				actual[i] = s
			}
		default:
			// Treat single value as a single-element array
			actual = []interface{}{v}
		}

		// Helper to check if actual array contains a string representation of any expected value
		contains := func(expected string) bool {
			for _, item := range actual {
				if fmt.Sprintf("%v", item) == expected {
					return true
				}
			}
			return false
		}

		switch rule.Op {
		case "==", "in":
			// For "in" or "==", we split rule.Value by commas and check if any expected value is contained in the actual array
			parts := strings.Split(rule.Value, ",")
			for _, p := range parts {
				if contains(strings.TrimSpace(p)) {
					return true
				}
			}
			return false
		case "!=":
			parts := strings.Split(rule.Value, ",")
			for _, p := range parts {
				if contains(strings.TrimSpace(p)) {
					return false
				}
			}
			return true
		}
	case "integer":
		var actual int64
		switch v := val.(type) {
		case int:
			actual = int64(v)
		case int32:
			actual = int64(v)
		case int64:
			actual = v
		case float64:
			actual = int64(v)
		case float32:
			actual = int64(v)
		case string:
			i, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return false
			}
			actual = i
		default:
			return false
		}

		switch rule.Op {
		case ">":
			expect, err := strconv.ParseInt(rule.Value, 10, 64)
			if err != nil {
				return false
			}
			return actual > expect
		case ">=":
			expect, err := strconv.ParseInt(rule.Value, 10, 64)
			if err != nil {
				return false
			}
			return actual >= expect
		case "==":
			expect, err := strconv.ParseInt(rule.Value, 10, 64)
			if err != nil {
				return false
			}
			return actual == expect
		case "<":
			expect, err := strconv.ParseInt(rule.Value, 10, 64)
			if err != nil {
				return false
			}
			return actual < expect
		case "<=":
			expect, err := strconv.ParseInt(rule.Value, 10, 64)
			if err != nil {
				return false
			}
			return actual <= expect
		case "!=":
			expect, err := strconv.ParseInt(rule.Value, 10, 64)
			if err != nil {
				return false
			}
			return actual != expect
		case "in":
			parts := strings.Split(rule.Value, ",")
			for _, p := range parts {
				if exp, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64); err == nil && actual == exp {
					return true
				}
			}
			return false
		}
	case "float":
		var actual float64
		switch v := val.(type) {
		case float64:
			actual = v
		case float32:
			actual = float64(v)
		case int:
			actual = float64(v)
		case int64:
			actual = float64(v)
		case string:
			f, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return false
			}
			actual = f
		default:
			return false
		}

		switch rule.Op {
		case ">":
			expect, err := strconv.ParseFloat(rule.Value, 64)
			if err != nil {
				return false
			}
			return actual > expect
		case ">=":
			expect, err := strconv.ParseFloat(rule.Value, 64)
			if err != nil {
				return false
			}
			return actual >= expect
		case "==":
			expect, err := strconv.ParseFloat(rule.Value, 64)
			if err != nil {
				return false
			}
			return actual == expect
		case "<":
			expect, err := strconv.ParseFloat(rule.Value, 64)
			if err != nil {
				return false
			}
			return actual < expect
		case "<=":
			expect, err := strconv.ParseFloat(rule.Value, 64)
			if err != nil {
				return false
			}
			return actual <= expect
		case "!=":
			expect, err := strconv.ParseFloat(rule.Value, 64)
			if err != nil {
				return false
			}
			return actual != expect
		}
	case "boolean":
		var actual bool
		switch v := val.(type) {
		case bool:
			actual = v
		case string:
			b, err := strconv.ParseBool(v)
			if err != nil {
				return false
			}
			actual = b
		default:
			return false
		}

		expect, err := strconv.ParseBool(rule.Value)
		if err != nil {
			return false
		}

		switch rule.Op {
		case "==":
			return actual == expect
		case "!=":
			return actual != expect
		}
	case "string":
		actual := fmt.Sprintf("%v", val)
		expect := rule.Value

		switch rule.Op {
		case "==":
			return actual == expect
		case "!=":
			return actual != expect
		case "in":
			parts := strings.Split(rule.Value, ",")
			for _, p := range parts {
				if actual == strings.TrimSpace(p) {
					return true
				}
			}
			return false
		}
	}
	return false
}

// matchCondition 驱动执行整个规则级条件系统的递归评估入口。
func matchCondition(cond *Condition, params map[string]interface{}) bool {
	if cond == nil {
		return true
	}
	// 优先评估 AllOf 条件群
	if len(cond.AllOf) > 0 {
		for _, allOf := range cond.AllOf {
			if !matchBasicRule(allOf, params) {
				return false
			}
		}
		return true
	}
	// 评估 Item 条件群
	if len(cond.OneOf) > 0 {
		for _, anyOf := range cond.OneOf {
			if matchBasicRule(anyOf, params) {
				return true
			}
		}
		return false
	}
	return true
}
