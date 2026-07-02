package price_calcular

import (
	"github.com/Knetic/govaluate"
	"regexp"
)

type IExpression interface {
	String() string
	Expr() *govaluate.EvaluableExpression
}

type Expression struct {
	org  string
	expr *govaluate.EvaluableExpression
}

func (e *Expression) Expr() *govaluate.EvaluableExpression {
	return e.expr
}

func (e *Expression) String() string {
	return e.org
}

func NewExpression(expr string) (IExpression, error) {
	processedExpr := preProcessExpression(expr)
	evaluableExpr, err := govaluate.NewEvaluableExpression(processedExpr)
	if err != nil {
		return nil, err
	}
	return &Expression{
		org:  expr,
		expr: evaluableExpr,
	}, nil
}

// exprRegex 正则表达式用于匹配类似 #cost.per_second, #sale.input_token 价格属性前缀。
// 在传入 govaluate 前需要将前缀及点 "#cost." 整体替换为合法的变量命名（如 "cost_"），使其兼容 govaluate 词法。
var exprRegex = regexp.MustCompile(`#([a-zA-Z0-9]+)\.([a-zA-Z0-9_]+)`)

// preProcessExpression 将表达式中的 "#prefix.field" 替换为 "prefix_field" 格式，使其兼容 govaluate 词法。
func preProcessExpression(expr string) string {
	return exprRegex.ReplaceAllString(expr, "${1}_${2}")
}
