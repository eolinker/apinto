package amount

import (
	"context"
	"sync"
	"testing"
	"time"

	context_label "github.com/eolinker/apinto/common/context-label"
	quota_limiting_strategy "github.com/eolinker/apinto/drivers/strategy/quota-limiting-strategy"
	"github.com/eolinker/apinto/resources"
	"github.com/eolinker/apinto/utils/response"
	"github.com/eolinker/eosc/eocontext"
)

type dummyTestContext struct {
	eocontext.EoContext
	labels map[string]string
	values map[string]interface{}
}

func (d *dummyTestContext) GetLabel(key string) string {
	if d.labels == nil {
		return ""
	}
	return d.labels[key]
}

func (d *dummyTestContext) Value(key interface{}) interface{} {
	if k, ok := key.(string); ok && d.values != nil {
		return d.values[k]
	}
	return nil
}

func (d *dummyTestContext) WithValue(key, val interface{}) {
	if d.values == nil {
		d.values = make(map[string]interface{})
	}
	if k, ok := key.(string); ok {
		d.values[k] = val
	}
}

func (d *dummyTestContext) SetLabel(key, value string) {
	if d.labels == nil {
		d.labels = make(map[string]string)
	}
	d.labels[key] = value
}

func (d *dummyTestContext) Context() context.Context {
	return context.Background()
}

type dummyStrategy struct {
	id         string
	targetType string
	period     quota_limiting_strategy.Period
	threshold  int64
}

func (s *dummyStrategy) ID() string {
	return s.id
}

func (s *dummyStrategy) TargetType() string {
	return s.targetType
}

func (s *dummyStrategy) Period() quota_limiting_strategy.Period {
	return s.period
}

func (s *dummyStrategy) Threshold() int64 {
	return s.threshold
}

func (s *dummyStrategy) Response() response.IResponse {
	return nil
}

// mockCache 模拟内存 Cache
type mockCache struct {
	resources.ICache
	mu   sync.Mutex
	data map[string]int64
}

func newMockCache() *mockCache {
	return &mockCache{data: make(map[string]int64)}
}

func (m *mockCache) IncrBy(ctx context.Context, key string, amount int64, expiration time.Duration) resources.IntResult {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] += amount
	return resources.NewIntResult(m.data[key], nil)
}

func (m *mockCache) DecrBy(ctx context.Context, key string, amount int64, expiration time.Duration) resources.IntResult {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] -= amount
	return resources.NewIntResult(m.data[key], nil)
}

func TestBuildQuotaKeyAndTTL(t *testing.T) {
	ctx := &dummyTestContext{
		labels: map[string]string{
			"product":     "xpack",
			"consumer":    "user_uuid_123",
			"tenant":      "tenant_uuid_456",
			"application": "app_789",
		},
	}

	stUser := &dummyStrategy{
		id:         "strat_uuid_abc",
		targetType: "user",
		period:     quota_limiting_strategy.PeriodHour,
		threshold:  100,
	}

	st := &Strategy{
		key: context_label.NewKeyGenerator("{product}:quota-limiting:{strategy}:{target_type}:{application}:{period}:{time_format}"),
	}

	now := time.Date(2027, 8, 5, 15, 4, 5, 0, time.UTC)

	key, ttl := st.buildQuotaKeyAndTTL(ctx, stUser, now)
	expectedKey := "xpack:quota-limiting:strat_uuid_abc:user:app_789:hour:2027080515"
	if key != expectedKey {
		t.Fatalf("expected key %s, got %s", expectedKey, key)
	}
	if ttl <= 0 {
		t.Fatalf("expected positive ttl, got %v", ttl)
	}
}

func TestMockCachePreDeductAndRollback(t *testing.T) {
	mc := newMockCache()
	ctx := context.Background()

	// 1. 预扣 1000000 (代表1.0元)
	mc.IncrBy(ctx, "k1", 1000000, 0)
	if mc.data["k1"] != 1000000 {
		t.Fatalf("expected 1000000, got %d", mc.data["k1"])
	}

	// 2. 超额回滚 1000000
	mc.DecrBy(ctx, "k1", 1000000, 0)
	if mc.data["k1"] != 0 {
		t.Fatalf("expected 0, got %d", mc.data["k1"])
	}

	// 3. 补扣（差额为 +200000）
	mc.IncrBy(ctx, "k1", 1000000, 0) // 先预扣 1000000
	mc.IncrBy(ctx, "k1", 200000, 0)  // 实际 1200000，补扣 200000
	if mc.data["k1"] != 1200000 {
		t.Fatalf("expected 1200000, got %d", mc.data["k1"])
	}
}
