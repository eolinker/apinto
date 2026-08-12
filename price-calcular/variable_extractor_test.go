package price_calcular

import (
	"testing"

	"github.com/tidwall/gjson"
)

// TestConvertType 测试从原始对象到目标数值类型的自动无损转换
func TestConvertType(t *testing.T) {
	// 1. 转换为 integer
	v1, err := convertRawType("123", "integer")
	if err != nil || v1.(int64) != 123 {
		t.Errorf("failed string -> integer: %v (err: %v)", v1, err)
	}

	v2, err := convertRawType(45.67, "integer")
	if err != nil || v2.(int64) != 45 {
		t.Errorf("failed float64 -> integer: %v (err: %v)", v2, err)
	}

	// 2. 转换为 float
	v3, err := convertRawType("3.14159", "float")
	if err != nil || v3.(float64) != 3.14159 {
		t.Errorf("failed string -> float: %v (err: %v)", v3, err)
	}

	v4, err := convertRawType(100, "float")
	if err != nil || v4.(float64) != 100.0 {
		t.Errorf("failed int -> float: %v (err: %v)", v4, err)
	}

	// 3. 转换为 boolean
	v5, err := convertRawType("true", "boolean")
	if err != nil || v5.(bool) != true {
		t.Errorf("failed string -> boolean: %v (err: %v)", v5, err)
	}

	v6, err := convertRawType("1", "boolean")
	if err != nil || v6.(bool) != true {
		t.Errorf("failed numeric string -> boolean: %v (err: %v)", v6, err)
	}

	// 4. 转换为 string
	v7, err := convertRawType(12345.6, "string")
	if err != nil || v7.(string) != "12345.6" {
		t.Errorf("failed any -> string: %v (err: %v)", v7, err)
	}
}

// TestNewVariablesExtractor_ConfigErrors 测试错误或不支持的配置源的校验能力
func TestNewVariablesExtractor_ConfigErrors(t *testing.T) {
	invalidVars := map[string]*Variable{
		"my_var": {
			Source: "invalid_source", // 错误配置
			Type:   "integer",
		},
	}

	_, err := NewVariablesExtractor(invalidVars)
	if err == nil {
		t.Error("expected error for unsupported source 'invalid_source', but got nil")
	}
}

// TestBodyExtractor_ExtractFromChunk_SSE 测试流式提取器对标准 SSE (data:) 格式帧的极速解析提取能力
func TestBodyExtractor_ExtractFromChunk_SSE(t *testing.T) {
	extractor, err := NewBodyExtractor("choices.0.delta.content", false, "string")
	if err != nil {
		t.Fatalf("failed to create body extractor: %v", err)
	}

	// 模拟 SSE 流式数据帧（单 chunk 内合并包含两帧，并伴随 event 标签和 DONE 标识）
	sseChunk := []byte(`
event: message_delta
data: {"choices":[{"delta":{"content":"Hello"}}]}

event: message_delta
data: {"choices":[{"delta":{"content":" world!"}}]}
data: [DONE]
`)

	val, err := extractor.ExtractFromChunk(sseChunk)
	if err != nil {
		t.Fatalf("extract stream chunk failed: %v", err)
	}

	// 我们提取的应当是最后一帧 data 的有效数据
	expected := " world!"
	if val.(string) != expected {
		t.Errorf("expected extracted string %q, got '%v'", expected, val)
	}
}

// TestBodyExtractor_ExtractFromChunk_Fallback 测试流式提取器对不带 data: 标签普通 json 数据块的降级解析匹配
func TestBodyExtractor_ExtractFromChunk_Fallback(t *testing.T) {
	extractor, err := NewBodyExtractor("usage.total_tokens", false, "integer")
	if err != nil {
		t.Fatalf("failed to create extractor: %v", err)
	}

	// 模拟流式结束时发送的一个纯 json (没有 data: 头) 用量包
	rawJsonChunk := []byte(`{"usage":{"total_tokens": 150}}`)

	val, err := extractor.ExtractFromChunk(rawJsonChunk)
	if err != nil {
		t.Fatalf("extract chunk failed: %v", err)
	}

	var expected int64 = 150
	if val.(int64) != expected {
		t.Errorf("expected extracted integer %v, got %v", expected, val)
	}
}

func TestConvertJSONPathToGjson(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"$.content[?(@.role==\"reference_video\")]", "content.#(role==\"reference_video\")"},
		{"content[?(@.role==\"reference_video\")]", "content.#(role==\"reference_video\")"},
		{"$.choices.0.message.content", "choices.0.message.content"},
	}

	for _, tt := range tests {
		actual := convertJSONPathToGjson(tt.input)
		if actual != tt.expected {
			t.Errorf("convertJSONPathToGjson(%q) = %q, expected %q", tt.input, actual, tt.expected)
		}
	}
}

func TestBodyExtractor_BooleanExistence(t *testing.T) {
	extractor, err := NewBodyExtractor("$.content[?(@.role==\"reference_video\")]", false, "boolean")
	if err != nil {
		t.Fatalf("failed to create body extractor: %v", err)
	}

	// 1. Matches exist
	bodyWithVideo := []byte(`{"content": [{"role": "user"}, {"role": "reference_video"}]}`)
	res := gjson.GetBytes(bodyWithVideo, extractor.path)
	val, err := convertGjsonType(res, extractor.varType)
	if err != nil || val.(bool) != true {
		t.Errorf("expected boolean existence to be true, got %v (err: %v)", val, err)
	}

	// 2. Matches do not exist
	bodyWithoutVideo := []byte(`{"content": [{"role": "user"}]}`)
	res2 := gjson.GetBytes(bodyWithoutVideo, extractor.path)
	val2, err := convertGjsonType(res2, extractor.varType)
	if err != nil || val2.(bool) != false {
		t.Errorf("expected boolean existence to be false, got %v (err: %v)", val2, err)
	}
}

func TestBodyExtractor_ArrayType(t *testing.T) {
	extractor, err := NewBodyExtractor("$.content.#.role", false, "array")
	if err != nil {
		t.Fatalf("failed to create body extractor: %v", err)
	}

	body := []byte(`{"content": [{"role": "user"}, {"role": "reference_video"}]}`)
	res := gjson.GetBytes(body, extractor.path)
	val, err := convertGjsonType(res, extractor.varType)
	if err != nil {
		t.Fatalf("convert array type error: %v", err)
	}

	arr, ok := val.([]interface{})
	if !ok || len(arr) != 2 || arr[0] != "user" || arr[1] != "reference_video" {
		t.Errorf("expected [\"user\", \"reference_video\"], got %v", val)
	}
}

func TestMatchBasicRule_Array(t *testing.T) {
	rule1 := &BasicRule{
		Key:   "roles",
		Op:    "in",
		Value: "reference_video,admin",
		Type:  "array",
	}

	rule2 := &BasicRule{
		Key:   "roles",
		Op:    "==",
		Value: "admin",
		Type:  "array",
	}

	params := map[string]interface{}{
		"roles": []interface{}{"user", "reference_video"},
	}

	if !matchBasicRule(rule1, params) {
		t.Error("expected rule1 (in) to match")
	}

	if matchBasicRule(rule2, params) {
		t.Error("expected rule2 (== admin) not to match")
	}
}

func TestBodyExtractor_CandidatesTokensDetails(t *testing.T) {
	// 1. 验证 JSONPath 转换
	inputPath := `$.usageMetadata.candidatesTokensDetails[?(@.modality=="IMAGE")].tokenCount`
	expectedConverted := `usageMetadata.candidatesTokensDetails.#(modality=="IMAGE").tokenCount`
	actualConverted := convertJSONPathToGjson(inputPath)
	if actualConverted != expectedConverted {
		t.Errorf("convertJSONPathToGjson(%q) = %q, expected %q", inputPath, actualConverted, expectedConverted)
	}

	// 2. 验证提取器是否能正确提取目标数据
	extractor, err := NewBodyExtractor(inputPath, false, "integer")
	if err != nil {
		t.Fatalf("failed to create body extractor: %v", err)
	}

	rawJsonChunk := []byte(`{
		"usageMetadata": { 
			"promptTokenCount": 16, 
			"candidatesTokenCount": 1541, 
			"totalTokenCount": 1557, 
			"promptTokensDetails": [ 
				{ 
					"modality": "TEXT", 
					"tokenCount": 16 
				} 
			], 
			"candidatesTokensDetails": [ 
				{ 
					"modality": "IMAGE", 
					"tokenCount": 1120 
				} 
			], 
			"serviceTier": "standard" 
		}
	}`)

	val, err := extractor.ExtractFromChunk(rawJsonChunk)
	if err != nil {
		t.Fatalf("extract chunk failed: %v", err)
	}

	expectedVal := int64(1120)
	if val.(int64) != expectedVal {
		t.Errorf("expected extracted integer %v, got %v", expectedVal, val)
	}
}
