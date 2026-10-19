package price_calcular

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/Knetic/govaluate"
)

type IExpression interface {
	String() string
	Expr() *govaluate.EvaluableExpression
	NumExpr(params map[string]interface{}) string
}

type Expression struct {
	org       string
	processed string
	expr      *govaluate.EvaluableExpression
}

func (e *Expression) Expr() *govaluate.EvaluableExpression {
	return e.expr
}

func (e *Expression) String() string {
	return e.org
}

func (e *Expression) NumExpr(params map[string]interface{}) string {
	if e == nil || e.processed == "" {
		return ""
	}
	return varIdentRegex.ReplaceAllStringFunc(e.processed, func(name string) string {
		if name == "true" || name == "false" {
			return name
		}
		if params == nil {
			return "0"
		}
		val, ok := params[name]
		if !ok || val == nil {
			return "0"
		}
		return formatParamValue(val)
	})
}

func formatParamValue(val interface{}) string {
	switch v := val.(type) {
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case int:
		return strconv.Itoa(v)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case string:
		if v == "" {
			return "0"
		}
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

func NewExpression(expr string) (IExpression, error) {
	processedExpr := preProcessExpression(expr)
	evaluableExpr, err := govaluate.NewEvaluableExpression(processedExpr)
	if err != nil {
		return nil, err
	}
	return &Expression{
		org:       expr,
		processed: processedExpr,
		expr:      evaluableExpr,
	}, nil
}

// exprRegex 正则表达式用于匹配类似 #cost.per_second, #sale.input_token 价格属性前缀。
// 在传入 govaluate 前需要将前缀及点 "#cost." 整体替换为合法的变量命名（如 "cost_"），使其兼容 govaluate 词法。
var exprRegex = regexp.MustCompile(`#([a-zA-Z0-9]+)\.([a-zA-Z0-9_]+)`)

// varIdentRegex 用于匹配预处理后表达式中的变量标识符，将其替换为实际数值生成数字表达式。
var varIdentRegex = regexp.MustCompile(`\b[a-zA-Z_][a-zA-Z0-9_]*\b`)

// preProcessExpression 将表达式中的 "#prefix.field" 替换为 "prefix_field" 格式，使其兼容 govaluate 词法。
func preProcessExpression(expr string) string {
	return exprRegex.ReplaceAllString(expr, "${1}_${2}")
}
