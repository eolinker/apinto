package failover_strategy

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/eolinker/eosc/eocontext"
)

type mockEoContext struct {
	eocontext.EoContext
	labels map[string]string
}

func newMockEoContext(labels map[string]string) *mockEoContext {
	return &mockEoContext{labels: labels}
}

func (m *mockEoContext) Assert(i interface{}) error {
	return errors.New("not http context")
}

func (m *mockEoContext) SetLabel(name, value string) {
	if m.labels == nil {
		m.labels = make(map[string]string)
	}
	m.labels[name] = value
}

func (m *mockEoContext) GetLabel(name string) string {
	if m.labels == nil {
		return ""
	}
	return m.labels[name]
}

func (m *mockEoContext) Labels() map[string]string {
	return m.labels
}

func TestBaseConfig_Serialization(t *testing.T) {
	// 测试 api_key/base_url 与 apikey/baseUrl 互相兼容
	rawJSON := `{"apikey":"test-key","baseUrl":"https://api.openai.com/v1"}`
	var cfg BaseConfig
	err := json.Unmarshal([]byte(rawJSON), &cfg)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if cfg.APIKey != "test-key" || cfg.BaseUrl != "https://api.openai.com/v1" {
		t.Fatalf("unexpected fields: %+v", cfg)
	}

	marshaled, err := json.Marshal(&cfg)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	var outMap map[string]interface{}
	_ = json.Unmarshal(marshaled, &outMap)
	if outMap["api_key"] != "test-key" || outMap["apikey"] != "test-key" {
		t.Fatalf("expected api_key/apikey in json: %s", string(marshaled))
	}
	if outMap["base_url"] != "https://api.openai.com/v1" || outMap["baseUrl"] != "https://api.openai.com/v1" {
		t.Fatalf("expected base_url/baseUrl in json: %s", string(marshaled))
	}
}

func TestExtractor_ExactAndFallback(t *testing.T) {
	ext := NewExtractor("provider")

	h1, _ := NewHandler(&Config{
		Name:     "strat-exact-openai",
		Priority: 10,
	})
	h2, _ := NewHandler(&Config{
		Name:     "strat-wildcard",
		Priority: 5,
	})

	// 注册精确值与通配符
	ext.Set("strat-1", []string{"openai"}, []IHandler{h1})
	ext.Set("strat-2", []string{"*"}, []IHandler{h2})

	// 1. 访问 openai: 命中精确值 h1 和通配符兜底 h2
	ctx1 := newMockEoContext(map[string]string{"provider": "openai"})
	handlers, ok := ext.Get(ctx1)
	if !ok || len(handlers) != 2 {
		t.Fatalf("expected 2 handlers, got %d (ok=%v)", len(handlers), ok)
	}
	if handlers[0].Name() != "strat-exact-openai" || handlers[1].Name() != "strat-wildcard" {
		t.Fatalf("unexpected handlers order: %+v", handlers)
	}

	// 2. 访问 claude: 未命中精确值，仅命中通配符兜底 h2
	ctx2 := newMockEoContext(map[string]string{"provider": "claude"})
	handlers2, ok := ext.Get(ctx2)
	if !ok || len(handlers2) != 1 {
		t.Fatalf("expected 1 fallback handler, got %d", len(handlers2))
	}
	if handlers2[0].Name() != "strat-wildcard" {
		t.Fatalf("expected strat-wildcard, got %s", handlers2[0].Name())
	}

	// 3. 删除 strat-1: openai 精确索引被清理，只能命中通配符
	ext.Del("strat-1")
	handlersAfterDel, ok := ext.Get(ctx1)
	if !ok || len(handlersAfterDel) != 1 || handlersAfterDel[0].Name() != "strat-wildcard" {
		t.Fatalf("expected only fallback handler after del, got %+v", handlersAfterDel)
	}
}
