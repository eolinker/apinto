package model_mapping

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/eolinker/apinto/strategy"
	"github.com/eolinker/eosc/eocontext"
)

type mockContext struct {
	labels map[string]string
}

func newMockContext(labels map[string]string) *mockContext {
	return &mockContext{labels: labels}
}

func (m *mockContext) RequestId() string                       { return "req-1" }
func (m *mockContext) AcceptTime() time.Time                   { return time.Now() }
func (m *mockContext) Context() context.Context                { return context.Background() }
func (m *mockContext) Value(key interface{}) interface{}       { return nil }
func (m *mockContext) WithValue(key, val interface{})          {}
func (m *mockContext) Scheme() string                          { return "http" }
func (m *mockContext) Assert(i interface{}) error              { return nil }
func (m *mockContext) IsCloneable() bool                       { return false }
func (m *mockContext) Clone() (eocontext.EoContext, error)     { return m, nil }
func (m *mockContext) SetLabel(name, value string)             { m.labels[name] = value }
func (m *mockContext) GetLabel(name string) string             { return m.labels[name] }
func (m *mockContext) Labels() map[string]string               { return m.labels }
func (m *mockContext) GetComplete() eocontext.CompleteHandler  { return nil }
func (m *mockContext) SetCompleteHandler(handler eocontext.CompleteHandler) {}
func (m *mockContext) GetFinish() eocontext.FinishHandler      { return nil }
func (m *mockContext) SetFinish(handler eocontext.FinishHandler) {}
func (m *mockContext) GetBalance() eocontext.BalanceHandler    { return nil }
func (m *mockContext) SetBalance(handler eocontext.BalanceHandler) {}
func (m *mockContext) GetUpstreamHostHandler() eocontext.UpstreamHostHandler { return nil }
func (m *mockContext) SetUpstreamHostHandler(handler eocontext.UpstreamHostHandler) {}
func (m *mockContext) LocalIP() net.IP                         { return net.IPv4(127, 0, 0, 1) }
func (m *mockContext) LocalAddr() net.Addr                     { return nil }
func (m *mockContext) LocalPort() int                          { return 8080 }
func (m *mockContext) RealIP() string                          { return "127.0.0.1" }

func TestFilterTree_Match(t *testing.T) {
	config := strategy.FilterConfig{
		"consumer": []string{"xxxx", "yyyy"},
		"model":    []string{"claude-opus-5"},
	}

	tree, err := NewFilterTree(config)
	if err != nil {
		t.Fatalf("NewFilterTree failed: %v", err)
	}

	tests := []struct {
		name     string
		labels   map[string]string
		expected bool
	}{
		{
			name: "全匹配成功 - consumer为xxxx",
			labels: map[string]string{
				"consumer": "xxxx",
				"model":    "claude-opus-5",
			},
			expected: true,
		},
		{
			name: "全匹配成功 - consumer为yyyy",
			labels: map[string]string{
				"consumer": "yyyy",
				"model":    "claude-opus-5",
			},
			expected: true,
		},
		{
			name: "匹配失败 - consumer不匹配",
			labels: map[string]string{
				"consumer": "zzzz",
				"model":    "claude-opus-5",
			},
			expected: false,
		},
		{
			name: "匹配失败 - model不匹配",
			labels: map[string]string{
				"consumer": "xxxx",
				"model":    "gpt-4o",
			},
			expected: false,
		},
		{
			name: "匹配失败 - 缺少consumer label",
			labels: map[string]string{
				"model": "claude-opus-5",
			},
			expected: false,
		},
		{
			name: "匹配失败 - 缺少model label",
			labels: map[string]string{
				"consumer": "xxxx",
			},
			expected: false,
		},
		{
			name:     "匹配失败 - 上下文为空label",
			labels:   map[string]string{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newMockContext(tt.labels)
			actual := tree.Match(ctx)
			if actual != tt.expected {
				t.Errorf("Match() for %s = %v, expected %v", tt.name, actual, tt.expected)
			}
		})
	}
}

func TestFilterTree_AllMatch(t *testing.T) {
	// 测试包含通配符*的情况
	config := strategy.FilterConfig{
		"consumer": []string{"*"},
		"model":    []string{"claude-opus-5"},
	}

	tree, err := NewFilterTree(config)
	if err != nil {
		t.Fatalf("NewFilterTree failed: %v", err)
	}

	ctxMatch := newMockContext(map[string]string{
		"consumer": "any-consumer-id",
		"model":    "claude-opus-5",
	})
	if !tree.Match(ctxMatch) {
		t.Errorf("expected match with wildcard consumer, got false")
	}

	ctxNoModel := newMockContext(map[string]string{
		"consumer": "any-consumer-id",
		"model":    "unknown-model",
	})
	if tree.Match(ctxNoModel) {
		t.Errorf("expected fail with wrong model, got true")
	}
}

func TestFilterTree_MatchSpecifiedSubTrees(t *testing.T) {
	// 配置包含 consumer, model, path 三个子树
	config := strategy.FilterConfig{
		"consumer": []string{"user_123"},
		"model":    []string{"gpt-4"},
		"path":     []string{"/v1/chat/completions"},
	}

	tree, err := NewFilterTree(config)
	if err != nil {
		t.Fatalf("NewFilterTree failed: %v", err)
	}

	// 1. 指定只匹配 consumer 子树：即使 model 和 path 不匹配甚至不存在，只要 consumer 匹配就应该成功
	ctx1 := newMockContext(map[string]string{
		"consumer": "user_123",
		"model":    "unmatched-model",
		"path":     "/unmatched-path",
	})
	if !tree.Match(ctx1, "consumer") {
		t.Errorf("expected Match with specified 'consumer' to be true, got false")
	}

	// 2. 指定只匹配 consumer 和 model 两个子树：两者都匹配时成功
	ctx2 := newMockContext(map[string]string{
		"consumer": "user_123",
		"model":    "gpt-4",
		"path":     "/unmatched-path", // path 不匹配但未指定，因此忽略
	})
	if !tree.Match(ctx2, "consumer", "model") {
		t.Errorf("expected Match with specified 'consumer' and 'model' to be true, got false")
	}

	// 3. 指定匹配 consumer 和 model，其中 model 不匹配：失败
	ctx3 := newMockContext(map[string]string{
		"consumer": "user_123",
		"model":    "different-model",
	})
	if tree.Match(ctx3, "consumer", "model") {
		t.Errorf("expected Match with mismatched model to be false, got true")
	}

	// 4. 指定匹配树中未配置的 key（如 ip）：视为不设限跳过，只要其他指定的匹配即可
	ctx4 := newMockContext(map[string]string{
		"consumer": "user_123",
	})
	if !tree.Match(ctx4, "consumer", "ip") {
		t.Errorf("expected Match with non-configured 'ip' and matched 'consumer' to be true, got false")
	}

	// 5. 不传指定 keys 时，整树全匹配：由于 ctx2 中 path 不匹配，全匹配应返回 false
	if tree.Match(ctx2) {
		t.Errorf("expected full tree Match without keys to fail because path doesn't match, got true")
	}
}

func TestParseTarget(t *testing.T) {
	valid := "Anthropic/claude-opus-5"
	target, err := ParseTarget(valid)
	if err != nil {
		t.Fatalf("ParseTarget failed for valid input: %v", err)
	}
	if target.Provider != "Anthropic" || target.Model != "claude-opus-5" {
		t.Errorf("unexpected target parsed: %+v", target)
	}

	invalids := []string{
		"Anthropic",
		"Anthropic/",
		"/claude-opus-5",
		"Anthropic/claude/extra",
		"",
	}
	for _, inv := range invalids {
		if _, err := ParseTarget(inv); err == nil {
			t.Errorf("expected error for invalid target '%s', got nil", inv)
		}
	}
}
