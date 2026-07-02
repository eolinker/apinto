package context_label

import (
	"testing"

	http_context "github.com/eolinker/eosc/eocontext/http-context"
)

type mockHttpContext struct {
	http_context.IHttpContext
	labels map[string]string
}

func (m *mockHttpContext) GetLabel(key string) string {
	if m.labels == nil {
		return ""
	}
	return m.labels[key]
}

func TestKeyGenerator(t *testing.T) {
	tests := []struct {
		name     string
		template string
		labels   map[string]string
		expected string
	}{
		{
			name:     "simple",
			template: "balance:{application}",
			labels: map[string]string{
				"application": "app1",
			},
			expected: "balance:app1",
		},
		{
			name:     "nested",
			template: "{{product}:balance}:{balance_target}",
			labels: map[string]string{
				"product":        "apinto",
				"apinto:balance": "100",
				"balance_target": "usd",
			},
			expected: "100:usd",
		},
		{
			name:     "nested",
			template: "{{product}:balance}:{balance_target}",
			labels: map[string]string{
				"product":        "apinto",
				"balance_target": "usd",
			},
			expected: "{apinto:balance}:usd",
		},
		{
			name:     "partially_unresolved",
			template: "{{product}:balance}:{balance_target}",
			labels: map[string]string{
				"balance_target": "usd",
			},
			expected: "{{product}:balance}:usd",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kg := NewKeyGenerator(tt.template)
			ctx := &mockHttpContext{labels: tt.labels}
			actual := kg.Key(ctx)
			if actual != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, actual)
			}
		})
	}
}
