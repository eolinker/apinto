package billing

import (
	"testing"
)

func TestFieldExtractor_RequestExtract(t *testing.T) {
	fe, err := NewFieldExtractor(map[string]string{
		"model":  "$.model",
		"stream": "$.stream",
	}, nil)
	if err != nil {
		t.Fatalf("new extractor err: %v", err)
	}

	body := []byte(`{"model":"gpt-4","stream":true,"messages":[{"role":"user"}]}`)
	got, err := fe.ExtractFromRequest(body)
	if err != nil {
		t.Fatalf("extract err: %v", err)
	}
	if got["model"] != "gpt-4" {
		t.Fatalf("model want gpt-4, got %v", got["model"])
	}
	if got["stream"] != true {
		t.Fatalf("stream want true, got %v", got["stream"])
	}
}

func TestFieldExtractor_ResponseExtract(t *testing.T) {
	fe, err := NewFieldExtractor(nil, map[string]string{
		"input_count":  "$.usage.prompt_tokens",
		"output_count": "$.usage.completion_tokens",
		"task_id":      "$.id",
	})
	if err != nil {
		t.Fatalf("new extractor err: %v", err)
	}

	body := []byte(`{"id":"chatcmpl-9","usage":{"prompt_tokens":12,"completion_tokens":34}}`)
	got, err := fe.ExtractFromResponse(body)
	if err != nil {
		t.Fatalf("extract err: %v", err)
	}
	if got["input_count"].(int64) != 12 {
		t.Fatalf("input_count want 12, got %v", got["input_count"])
	}
	if got["output_count"].(int64) != 34 {
		t.Fatalf("output_count want 34, got %v", got["output_count"])
	}
	if got["task_id"] != "chatcmpl-9" {
		t.Fatalf("task_id want chatcmpl-9, got %v", got["task_id"])
	}
}

func TestFieldExtractor_EmptyExprs(t *testing.T) {
	fe, _ := NewFieldExtractor(nil, nil)

	got, err := fe.ExtractFromRequest([]byte(`bad-json`))
	if err != nil {
		t.Fatalf("empty exprs should not parse body, got err: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("want empty map, got %+v", got)
	}
}

func TestFieldExtractor_InvalidJSONPath(t *testing.T) {
	_, err := NewFieldExtractor(map[string]string{"x": "@@@invalid"}, nil)
	if err == nil {
		t.Fatalf("expect error for invalid jsonpath")
	}
}

func TestFieldExtractor_InvalidJSONBody(t *testing.T) {
	fe, _ := NewFieldExtractor(map[string]string{"x": "$.x"}, nil)
	_, err := fe.ExtractFromRequest([]byte(`{not-json`))
	if err == nil {
		t.Fatalf("expect parse err for malformed body")
	}
}

func TestFieldExtractor_MissingPathReturnsAbsent(t *testing.T) {
	fe, _ := NewFieldExtractor(map[string]string{
		"missing": "$.not_exist",
		"present": "$.foo",
	}, nil)
	got, err := fe.ExtractFromRequest([]byte(`{"foo":"bar"}`))
	if err != nil {
		t.Fatalf("extract err: %v", err)
	}
	if _, ok := got["missing"]; ok {
		t.Fatalf("missing path should not appear in result")
	}
	if got["present"] != "bar" {
		t.Fatalf("present want bar, got %v", got["present"])
	}
}
