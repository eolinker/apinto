package pricing_policy

import (
	"testing"
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
