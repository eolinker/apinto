package pricing

// ResultMatch 响应结果匹配规则，用于判断 API 响应是否进入计费流程。
// 只有响应满足全部已配置维度时才会执行 Calculate；不匹配则跳过计费。
type ResultMatch struct {
	StatusCodes  []int             `json:"status_codes,omitempty"`
	SuccessField string            `json:"success_field,omitempty"`
	SuccessValue interface{}       `json:"success_value,omitempty"`
	FieldRules   []ResultFieldRule `json:"field_rules,omitempty"`
}

// ResultFieldRule 单个字段匹配规则。Operator 为空或为 "==" / "=" 时按精确匹配，
// 否则按数值阈值（<=, >, >=, <, ==）比较。
type ResultFieldRule struct {
	Field    string      `json:"field"`
	Value    interface{} `json:"value"`
	Operator string      `json:"operator,omitempty"`
}
