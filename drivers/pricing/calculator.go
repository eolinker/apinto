package pricing_driver

import (
	"fmt"
	"strings"

	"github.com/eolinker/apinto/pricing"
)

// PriceCalculator 通用价格计算器，实现 pricing.IPriceCalculator 接口。
//
// 计算路径：
//  1. 同步/异步阶段决定 ChargeType（submit→pre，其余→actual）
//  2. PricingModeAdvanced 且存在规则时走 calculateAdvanced；否则走 calculateStandard
//
// 设计说明（与旧 ai-convert 版本的差异）：
//   - 资源维度：Model → Resource，移除 AI 专属字段
//   - 条件类型：移除 total_tokens/video_size 旧别名
//   - 新增条件类型：method（HTTP 方法）、path（HTTP path）
type PriceCalculator struct {
	provider       string
	pricing        *pricing.ResourcePricing
	phase          pricing.PricingPhase
	requestFields  map[string]string
	responseFields map[string]string
}

// NewPriceCalculator 构造一个新的计算器实例。
func NewPriceCalculator(provider string, p pricing.ResourcePricing, phase pricing.PricingPhase, requestFields, responseFields map[string]string) *PriceCalculator {
	return &PriceCalculator{
		provider:       provider,
		pricing:        &p,
		phase:          phase,
		requestFields:  requestFields,
		responseFields: responseFields,
	}
}

func (pc *PriceCalculator) Provider() string                  { return pc.provider }
func (pc *PriceCalculator) Resource() string                  { return pc.pricing.Resource }
func (pc *PriceCalculator) Phase() pricing.PricingPhase       { return pc.phase }
func (pc *PriceCalculator) Pricing() *pricing.ResourcePricing { return pc.pricing }
func (pc *PriceCalculator) RequestFields() map[string]string  { return pc.requestFields }
func (pc *PriceCalculator) ResponseFields() map[string]string { return pc.responseFields }

// Calculate 根据用量计算费用。
func (pc *PriceCalculator) Calculate(usage *pricing.PricingUsage) (*pricing.CalculateResult, error) {
	if usage == nil {
		return nil, fmt.Errorf("usage is nil")
	}
	chargeType := pricing.ChargeTypeActual
	if pc.phase == pricing.PricingPhaseSubmit {
		chargeType = pricing.ChargeTypePre
	}

	if pc.pricing.Mode == pricing.PricingModeAdvanced && len(pc.pricing.AdvancedRules) > 0 {
		return pc.calculateAdvanced(usage, chargeType)
	}
	return pc.calculateStandard(usage, chargeType)
}

// calculateStandard 标准计费：
//   - PricingModeToken：input/output/cached 按"每百万单位"
//   - PricingModePerCall：per_call × call_count
//   - PricingModePerSecond：per_second × duration
//   - PricingModeAdvanced 但无规则：等同 token 模式回退（输入输出按基础单价）
func (pc *PriceCalculator) calculateStandard(usage *pricing.PricingUsage, chargeType pricing.ChargeType) (*pricing.CalculateResult, error) {
	result := &pricing.CalculateResult{
		Mode:       pc.pricing.Mode,
		Phase:      pc.phase,
		ChargeType: chargeType,
	}

	switch pc.pricing.Mode {
	case pricing.PricingModeToken, pricing.PricingModeAdvanced:
		// PricingModeAdvanced 但 AdvancedRules 为空时回退到基础 token 公式
		result.InputCost = pc.pricing.Input * float64(usage.InputCount) / 1_000_000
		result.OutputCost = pc.pricing.Output * float64(usage.OutputCount) / 1_000_000
		result.CachedCost = pc.pricing.CachedInput * float64(usage.CachedCount) / 1_000_000
		result.TotalCost = result.InputCost + result.OutputCost + result.CachedCost

	case pricing.PricingModePerCall:
		callCount := usage.CallCount
		if callCount == 0 {
			callCount = 1
		}
		result.CallCost = pc.pricing.PerCall * float64(callCount)
		result.TotalCost = result.CallCost

	case pricing.PricingModePerSecond:
		result.DurationCost = pc.pricing.PerSecond * usage.Duration
		result.TotalCost = result.DurationCost

	default:
		return nil, fmt.Errorf("unsupported pricing mode: %s", pc.pricing.Mode)
	}

	return result, nil
}

// calculateAdvanced 高级阶梯计费：遍历规则首条命中后按其 TierPricing 计算；
// 无任何命中则回退到基础单价。
func (pc *PriceCalculator) calculateAdvanced(usage *pricing.PricingUsage, chargeType pricing.ChargeType) (*pricing.CalculateResult, error) {
	result := &pricing.CalculateResult{
		Mode:       pricing.PricingModeAdvanced,
		Phase:      pc.phase,
		ChargeType: chargeType,
	}

	matched := false
	for _, rule := range pc.pricing.AdvancedRules {
		if matchCondition(&rule.Condition, usage) {
			matched = true
			tier := rule.Pricing
			result.InputCost = tier.Input * float64(usage.InputCount) / 1_000_000
			result.OutputCost = tier.Output * float64(usage.OutputCount) / 1_000_000
			result.CachedCost = tier.CachedInput * float64(usage.CachedCount) / 1_000_000
			callCount := usage.CallCount
			if callCount == 0 {
				callCount = 1
			}
			result.CallCost = tier.PerCall * float64(callCount)
			result.DurationCost = tier.PerSecond * usage.Duration
			result.TotalCost = result.InputCost + result.OutputCost + result.CachedCost + result.CallCost + result.DurationCost
			break
		}
	}

	if !matched {
		result.InputCost = pc.pricing.Input * float64(usage.InputCount) / 1_000_000
		result.OutputCost = pc.pricing.Output * float64(usage.OutputCount) / 1_000_000
		result.CachedCost = pc.pricing.CachedInput * float64(usage.CachedCount) / 1_000_000
		callCount := usage.CallCount
		if callCount == 0 {
			callCount = 1
		}
		result.CallCost = pc.pricing.PerCall * float64(callCount)
		result.DurationCost = pc.pricing.PerSecond * usage.Duration
		result.TotalCost = result.InputCost + result.OutputCost + result.CachedCost + result.CallCost + result.DurationCost
	}

	return result, nil
}

// matchCondition 条件匹配。返回 true 表示命中阶梯规则。
//
// 支持的条件类型：
//   - total_count：按 usage.TotalCount 数值阈值比较
//   - input_type：字符串等值
//   - size：字符串等值
//   - status_code：按 usage.StatusCode 数值阈值比较
//   - method：按 usage.Method 字符串比较（大小写不敏感）
//   - path：按 usage.Path 字符串比较，支持后缀通配 "/foo/*"
//   - field：从 usage.Fields 取值，按值类型决定等值或阈值匹配
func matchCondition(cond *pricing.Condition, usage *pricing.PricingUsage) bool {
	switch cond.Type {
	case pricing.ConditionTotalCount:
		return compareThreshold(float64(usage.TotalCount), cond.Operator, cond.Threshold)
	case pricing.ConditionInputType:
		return usage.InputType == cond.Value
	case pricing.ConditionSize:
		return usage.Size == cond.Value
	case pricing.ConditionStatusCode:
		return compareThreshold(float64(usage.StatusCode), cond.Operator, cond.Threshold)
	case pricing.ConditionMethod:
		return strings.EqualFold(usage.Method, cond.Value)
	case pricing.ConditionPath:
		return matchPath(usage.Path, cond.Value)
	case pricing.ConditionField:
		fieldValue, ok := usage.Fields[cond.Field]
		if !ok {
			return false
		}
		return matchFieldCondition(fieldValue, cond)
	default:
		return false
	}
}

// matchPath 支持精确匹配和后缀通配："/foo/*" 命中所有以 /foo/ 开头的 path。
func matchPath(path, pattern string) bool {
	if pattern == "" {
		return false
	}
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(path, prefix)
	}
	return path == pattern
}

func matchFieldCondition(fieldValue interface{}, cond *pricing.Condition) bool {
	switch v := fieldValue.(type) {
	case string:
		return v == cond.Value
	case float64:
		return compareThreshold(v, cond.Operator, cond.Threshold)
	case int:
		return compareThreshold(float64(v), cond.Operator, cond.Threshold)
	case int64:
		return compareThreshold(float64(v), cond.Operator, cond.Threshold)
	default:
		return fmt.Sprintf("%v", fieldValue) == cond.Value
	}
}

func compareThreshold(value float64, operator string, threshold float64) bool {
	switch operator {
	case "<=":
		return value <= threshold
	case ">":
		return value > threshold
	case ">=":
		return value >= threshold
	case "<":
		return value < threshold
	case "==", "=":
		return value == threshold
	default:
		return false
	}
}
