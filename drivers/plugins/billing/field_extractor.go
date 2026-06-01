package billing

import (
	"fmt"

	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"
)

// FieldExtractor 通用字段提取器，使用 JSONPath 表达式从请求/响应 JSON 中提取计费字段。
//
// 表达式在配置加载阶段一次性编译为 jp.Expr 复用，避免每次请求重复解析。
type FieldExtractor struct {
	requestExprs  map[string]jp.Expr
	responseExprs map[string]jp.Expr
}

// NewFieldExtractor 解析并构造提取器。任一字段路径非法都会返回错误。
func NewFieldExtractor(requestFields, responseFields map[string]string) (*FieldExtractor, error) {
	reqExprs, err := parseFieldMap(requestFields)
	if err != nil {
		return nil, fmt.Errorf("parse request fields: %w", err)
	}
	respExprs, err := parseFieldMap(responseFields)
	if err != nil {
		return nil, fmt.Errorf("parse response fields: %w", err)
	}
	return &FieldExtractor{
		requestExprs:  reqExprs,
		responseExprs: respExprs,
	}, nil
}

func parseFieldMap(fields map[string]string) (map[string]jp.Expr, error) {
	exprs := make(map[string]jp.Expr, len(fields))
	for name, path := range fields {
		expr, err := jp.ParseString(path)
		if err != nil {
			return nil, fmt.Errorf("field %s: invalid jsonpath %s: %w", name, path, err)
		}
		exprs[name] = expr
	}
	return exprs, nil
}

// ExtractFromRequest 从请求 JSON 中按 RequestFields 提取字段。
// 表达式集合为空时直接返回空 map，避免无谓的解析开销。
func (fe *FieldExtractor) ExtractFromRequest(body []byte) (map[string]interface{}, error) {
	if len(fe.requestExprs) == 0 {
		return map[string]interface{}{}, nil
	}
	obj, err := oj.Parse(body)
	if err != nil {
		return nil, fmt.Errorf("parse request body: %w", err)
	}
	return fe.extract(obj, fe.requestExprs), nil
}

// ExtractFromResponse 从响应 JSON 中按 ResponseFields 提取字段。
func (fe *FieldExtractor) ExtractFromResponse(body []byte) (map[string]interface{}, error) {
	if len(fe.responseExprs) == 0 {
		return map[string]interface{}{}, nil
	}
	obj, err := oj.Parse(body)
	if err != nil {
		return nil, fmt.Errorf("parse response body: %w", err)
	}
	return fe.extract(obj, fe.responseExprs), nil
}

func (fe *FieldExtractor) extract(obj interface{}, exprs map[string]jp.Expr) map[string]interface{} {
	result := make(map[string]interface{}, len(exprs))
	for name, expr := range exprs {
		values := expr.Get(obj)
		if len(values) > 0 {
			result[name] = values[0]
		}
	}
	return result
}
