package billing

import (
	"fmt"

	"github.com/eolinker/apinto/pricing"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
)

// resultMatcher 三维度判断响应是否进入计费流程：
//
//  1. statusCodes：响应状态码必须在白名单中（未配置则跳过此维度）
//  2. successField：响应中指定字段必须等于 successValue
//  3. fieldRules：每条规则的字段值必须满足对应运算符
//
// 任一维度失败即返回 false。matcher 自身为 nil 视为默认通过（直接计费）。
type resultMatcher struct {
	statusCodes  []int
	successField string
	successValue interface{}
	fieldRules   []pricing.ResultFieldRule
	extractor    *FieldExtractor
}

func newResultMatcher(match *pricing.ResultMatch, extractor *FieldExtractor) (*resultMatcher, error) {
	if match == nil {
		return nil, nil
	}
	return &resultMatcher{
		statusCodes:  match.StatusCodes,
		successField: match.SuccessField,
		successValue: match.SuccessValue,
		fieldRules:   match.FieldRules,
		extractor:    extractor,
	}, nil
}

// Match 三阶段匹配；任一失败即拦截。allFields 会在第二阶段被原地补充以避免重复 JSON 解析。
func (m *resultMatcher) Match(ctx http_context.IHttpContext, allFields map[string]interface{}) bool {
	if m == nil {
		return true
	}

	// 阶段 1：状态码白名单
	if len(m.statusCodes) > 0 {
		statusCode := ctx.Response().StatusCode()
		found := false
		for _, code := range m.statusCodes {
			if statusCode == code {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// 阶段 2：成功字段
	if m.successField != "" {
		if m.extractor != nil {
			if _, exists := allFields[m.successField]; !exists {
				respBody := ctx.Response().GetBody()
				respFields, err := m.extractor.ExtractFromResponse(respBody)
				if err == nil && respFields != nil {
					for k, v := range respFields {
						allFields[k] = v
					}
				}
			}
		}
		fieldValue, ok := allFields[m.successField]
		if !ok {
			return false
		}
		if !matchValue(fieldValue, m.successValue) {
			return false
		}
	}

	// 阶段 3：字段规则
	for _, rule := range m.fieldRules {
		fieldValue, ok := allFields[rule.Field]
		if !ok {
			return false
		}
		if !matchFieldRuleValue(fieldValue, rule) {
			return false
		}
	}

	return true
}

// matchValue 实际值与期望值按基础类型比较；类型不一致时退化为字符串相等比较。
func matchValue(actual interface{}, expected interface{}) bool {
	switch ev := expected.(type) {
	case string:
		as, ok := actual.(string)
		if !ok {
			return fmt.Sprintf("%v", actual) == ev
		}
		return as == ev
	case float64:
		return compareNumeric(actual, ev)
	case int:
		return compareNumeric(actual, float64(ev))
	case bool:
		ab, ok := actual.(bool)
		if !ok {
			return false
		}
		return ab == ev
	default:
		return fmt.Sprintf("%v", actual) == fmt.Sprintf("%v", expected)
	}
}

// matchFieldRuleValue Operator 为空 / "==" / "=" 时按精确匹配，其他运算符走数值阈值比较。
func matchFieldRuleValue(actual interface{}, rule pricing.ResultFieldRule) bool {
	if rule.Operator == "" || rule.Operator == "==" || rule.Operator == "=" {
		return matchValue(actual, rule.Value)
	}
	av, ok := toFloat(actual)
	if !ok {
		return false
	}
	ev, ok := toFloat(rule.Value)
	if !ok {
		return false
	}
	return resultCompareThreshold(av, rule.Operator, ev)
}

func resultCompareThreshold(value float64, operator string, threshold float64) bool {
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

func compareNumeric(actual interface{}, expected float64) bool {
	av, ok := toFloat(actual)
	if !ok {
		return false
	}
	return av == expected
}

func toFloat(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	default:
		return 0, false
	}
}
