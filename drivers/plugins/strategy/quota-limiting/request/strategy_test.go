package request

import (
	quota_limiting_strategy "github.com/eolinker/apinto/drivers/strategy/quota-limiting-strategy"
	"testing"
	"time"
	
	"github.com/eolinker/eosc/eocontext"
)

type dummyTestContext struct {
	eocontext.EoContext
	labels map[string]string
}

func (d *dummyTestContext) GetLabel(key string) string {
	if d.labels == nil {
		return ""
	}
	return d.labels[key]
}

type dummyStrategy struct {
	id         string
	targetType string
	period     int
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

func TestBuildQuotaKeyAndTTL(t *testing.T) {
	ctx := &dummyTestContext{
		labels: map[string]string{
			"product":  "xpack",
			"consumer": "user_uuid_123",
			"tenant":   "tenant_uuid_456",
		},
	}
	
	// 1. 测试 TargetType 为 user
	stUser := &dummyStrategy{
		id:         "strat_uuid_abc",
		targetType: "user",
		period:     2, // hour
		threshold:  100,
	}
	
	now := time.Date(2027, 8, 5, 15, 4, 5, 0, time.UTC)
	
	key, ttl := buildQuotaKeyAndTTL(ctx, stUser, now)
	expectedKey := "xpack:quota-limiting:strat_uuid_abc:user:user_uuid_123:request:2027080515"
	if key != expectedKey {
		t.Fatalf("expected key %s, got %s", expectedKey, key)
	}
	if ttl <= 0 {
		t.Fatalf("expected positive ttl, got %v", ttl)
	}
	
	// 2. 测试 TargetType 为 channel (租户)
	stChannel := &dummyStrategy{
		id:         "strat_uuid_channel",
		targetType: "channel",
		period:     2, // hour
		threshold:  500,
	}
	
	keyChannel, _ := buildQuotaKeyAndTTL(ctx, stChannel, now)
	expectedKeyChannel := "xpack:quota-limiting:strat_uuid_channel:channel:tenant_uuid_456:request:2027080515"
	if keyChannel != expectedKeyChannel {
		t.Fatalf("expected key %s, got %s", expectedKeyChannel, keyChannel)
	}
}
