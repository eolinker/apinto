package request

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	context_label "github.com/eolinker/apinto/common/context-label"
	quota_limiting_strategy "github.com/eolinker/apinto/drivers/strategy/quota-limiting-strategy"
	"github.com/eolinker/apinto/resources"
	"github.com/eolinker/apinto/utils/response"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

// ============================================================================
// Mock & Dummy Structures
// ============================================================================

// dummyHttpResponse 实现 http_service.IResponse 接口
type dummyHttpResponse struct {
	statusCode int
	statusStr  string
	header     http.Header
	body       []byte
}

func newDummyHttpResponse() *dummyHttpResponse {
	return &dummyHttpResponse{
		header: make(http.Header),
	}
}

func (r *dummyHttpResponse) SetStatus(code int, status string) {
	r.statusCode = code
	r.statusStr = status
}

func (r *dummyHttpResponse) SetHeader(key, value string) {
	r.header.Set(key, value)
}

func (r *dummyHttpResponse) SetBody(body []byte) {
	r.body = body
}

func (r *dummyHttpResponse) Header() http.Header {
	return r.header
}

func (r *dummyHttpResponse) Status() string {
	return r.statusStr
}

func (r *dummyHttpResponse) StatusCode() int {
	return r.statusCode
}

func (r *dummyHttpResponse) ProxyStatus() string {
	return r.statusStr
}

func (r *dummyHttpResponse) ProxyStatusCode() int {
	return r.statusCode
}

func (r *dummyHttpResponse) SetProxyStatus(code int, status string) {
	r.statusCode = code
	r.statusStr = status
}

func (r *dummyHttpResponse) GetBody() []byte {
	return r.body
}

func (r *dummyHttpResponse) BodyLen() int {
	return len(r.body)
}

func (r *dummyHttpResponse) ContentType() string {
	return r.header.Get("Content-Type")
}

func (r *dummyHttpResponse) ClearError() {}
func (r *dummyHttpResponse) HeaderReset() { r.header = make(http.Header) }
func (r *dummyHttpResponse) HeadersString() string { return "" }
func (r *dummyHttpResponse) IsBodyStream() bool { return false }
func (r *dummyHttpResponse) RemoteIP() string { return "127.0.0.1" }
func (r *dummyHttpResponse) RemoteAddr() string { return "127.0.0.1:8080" }
func (r *dummyHttpResponse) RemotePort() int { return 8080 }
func (r *dummyHttpResponse) ResponseError() error { return nil }
func (r *dummyHttpResponse) ResponseTime() time.Duration { return 0 }

// 补全 http_service.IResponse 其他未用到的接口空实现
func (r *dummyHttpResponse) Headers() http.Header                    { return r.header }
func (r *dummyHttpResponse) AddHeader(key, value string)             {}
func (r *dummyHttpResponse) DelHeader(key string)                    {}
func (r *dummyHttpResponse) Response()                               {}
func (r *dummyHttpResponse) ContentLength() int                      { return len(r.body) }
func (r *dummyHttpResponse) GetHeader(key string) string             { return r.header.Get(key) }
func (r *dummyHttpResponse) String() string                          { return string(r.body) }
func (r *dummyHttpResponse) SetResponseTime(duration time.Duration) {}
func (r *dummyHttpResponse) GetResponseTime() time.Duration         { return 0 }

// dummyHttpContext 实现 http_service.IHttpContext
type dummyHttpContext struct {
	labels map[string]string
	values map[string]interface{}
	resp   *dummyHttpResponse
	ctx    context.Context
}

func newDummyHttpContext() *dummyHttpContext {
	return &dummyHttpContext{
		labels: make(map[string]string),
		values: make(map[string]interface{}),
		resp:   newDummyHttpResponse(),
		ctx:    context.Background(),
	}
}

func (d *dummyHttpContext) GetLabel(key string) string {
	if d.labels == nil {
		return ""
	}
	return d.labels[key]
}

func (d *dummyHttpContext) SetLabel(key, value string) {
	if d.labels == nil {
		d.labels = make(map[string]string)
	}
	d.labels[key] = value
}

func (d *dummyHttpContext) Value(key interface{}) interface{} {
	if k, ok := key.(string); ok && d.values != nil {
		return d.values[k]
	}
	return nil
}

func (d *dummyHttpContext) WithValue(key, val interface{}) {
	if d.values == nil {
		d.values = make(map[string]interface{})
	}
	if k, ok := key.(string); ok {
		d.values[k] = val
	}
}

func (d *dummyHttpContext) Context() context.Context {
	if d.ctx == nil {
		return context.Background()
	}
	return d.ctx
}

func (d *dummyHttpContext) Response() http_service.IResponse {
	if d.resp == nil {
		d.resp = newDummyHttpResponse()
	}
	return d.resp
}

func (d *dummyHttpContext) GetEntry() eosc.IEntry {
	return nil
}

func (d *dummyHttpContext) Proxies() []http_service.IProxy {
	return nil
}

func (d *dummyHttpContext) ProxyClone() http_service.IRequest { return nil }
func (d *dummyHttpContext) SetProxy(proxy http_service.IRequest) {}
func (d *dummyHttpContext) AcceptTime() time.Time { return time.Now() }
func (d *dummyHttpContext) Scheme() string { return "http" }
func (d *dummyHttpContext) Assert(i interface{}) error {
	if p, ok := i.(**dummyHttpContext); ok {
		*p = d
		return nil
	}
	if p, ok := i.(*http_service.IHttpContext); ok {
		*p = d
		return nil
	}
	if p, ok := i.(*eocontext.EoContext); ok {
		*p = d
		return nil
	}
	return errors.New("not support")
}
func (d *dummyHttpContext) Labels() map[string]string { return d.labels }
func (d *dummyHttpContext) GetComplete() eocontext.CompleteHandler { return nil }
func (d *dummyHttpContext) SetCompleteHandler(handler eocontext.CompleteHandler) {}
func (d *dummyHttpContext) GetFinish() eocontext.FinishHandler { return nil }
func (d *dummyHttpContext) SetFinish(handler eocontext.FinishHandler) {}
func (d *dummyHttpContext) GetBalance() eocontext.BalanceHandler { return nil }
func (d *dummyHttpContext) SetBalance(handler eocontext.BalanceHandler) {}
func (d *dummyHttpContext) GetUpstreamHostHandler() eocontext.UpstreamHostHandler { return nil }
func (d *dummyHttpContext) SetUpstreamHostHandler(handler eocontext.UpstreamHostHandler) {}
func (d *dummyHttpContext) RealIP() string { return "127.0.0.1" }
func (d *dummyHttpContext) LocalIP() net.IP { return net.ParseIP("127.0.0.1") }
func (d *dummyHttpContext) LocalAddr() net.Addr { return nil }
func (d *dummyHttpContext) LocalPort() int { return 80 }
func (d *dummyHttpContext) IsCloneable() bool { return false }
func (d *dummyHttpContext) Clone() (eocontext.EoContext, error) { return d, nil }
func (d *dummyHttpContext) SendTo(scheme string, node eocontext.INode, timeout time.Duration) error { return nil }

// 补全 http_service.IHttpContext 其他接口的空实现
func (d *dummyHttpContext) Request() http_service.IRequestReader { return nil }
func (d *dummyHttpContext) Proxy() http_service.IRequest       { return nil }
func (d *dummyHttpContext) RequestId() string                  { return "req-123" }
func (d *dummyHttpContext) IsAccept() bool                     { return true }
func (d *dummyHttpContext) SetAccept(accept bool)              {}
func (d *dummyHttpContext) FastFinish()                        {}

// dummyChain 模拟责任链
type dummyChain struct {
	called bool
}

func (c *dummyChain) DoChain(ctx eocontext.EoContext) error {
	c.called = true
	return nil
}

func (c *dummyChain) Destroy() {}

// dummyCustomResponse 模拟自定义 Response
type dummyCustomResponse struct {
	called bool
}

func (r *dummyCustomResponse) Response(ctx eocontext.EoContext) {
	r.called = true
}

// dummyStrategy 模拟单个 IStrategy
type dummyStrategy struct {
	id         string
	targetType string
	period     quota_limiting_strategy.Period
	threshold  int64
	resp       *dummyCustomResponse
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
	if s.resp == nil {
		return nil
	}
	return s.resp
}

// mockCache 线程安全 Mock Cache
type mockCache struct {
	resources.ICache
	mu        sync.RWMutex
	data      map[string]int64
	decrCalls map[string]int64
	incrErr   error
}

func newMockCache() *mockCache {
	return &mockCache{
		data:      make(map[string]int64),
		decrCalls: make(map[string]int64),
	}
}

func (m *mockCache) IncrBy(ctx context.Context, key string, amount int64, expiration time.Duration) resources.IntResult {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.incrErr != nil {
		return resources.NewIntResult(0, m.incrErr)
	}
	m.data[key] += amount
	return resources.NewIntResult(m.data[key], nil)
}

func (m *mockCache) DecrBy(ctx context.Context, key string, amount int64, expiration time.Duration) resources.IntResult {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] -= amount
	m.decrCalls[key] += amount
	return resources.NewIntResult(m.data[key], nil)
}

func (m *mockCache) GetVal(key string) int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.data[key]
}

func (m *mockCache) GetDecrCalls(key string) int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.decrCalls[key]
}

// ============================================================================
// Unit Tests
// ============================================================================

func TestCheckConfigAndDriver(t *testing.T) {
	// Test CheckConfig
	cfgEmpty := &Config{}
	if err := CheckConfig(cfgEmpty, nil); err != nil {
		t.Fatalf("CheckConfig error: %v", err)
	}
	expectedDefaultKey := "{product}:quota-limiting:{strategy}:{target_type}:{application}:{period}:{time_format}"
	if cfgEmpty.Key != expectedDefaultKey {
		t.Fatalf("expected key %s, got %s", expectedDefaultKey, cfgEmpty.Key)
	}

	cfgCustom := &Config{Key: "custom_key"}
	if err := CheckConfig(cfgCustom, nil); err != nil {
		t.Fatalf("CheckConfig error: %v", err)
	}
	if cfgCustom.Key != "custom_key" {
		t.Fatalf("expected custom_key, got %s", cfgCustom.Key)
	}

	// Test Create
	worker, err := Create("id1", "name1", &Config{Cache: "cache_id"}, nil)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	st, ok := worker.(*Strategy)
	if !ok {
		t.Fatalf("expected *Strategy worker, got %T", worker)
	}
	if st.redisID != "cache_id" {
		t.Fatalf("expected redisID cache_id, got %s", st.redisID)
	}

	// Test Reset & Lifecycle
	if err := st.Reset(&Config{Cache: "new_cache_id"}, nil); err != nil {
		t.Fatalf("Reset error: %v", err)
	}
	if st.redisID != "new_cache_id" {
		t.Fatalf("expected new_cache_id, got %s", st.redisID)
	}

	if err := st.Start(); err != nil {
		t.Errorf("Start returned error: %v", err)
	}
	if err := st.Stop(); err != nil {
		t.Errorf("Stop returned error: %v", err)
	}
	st.Destroy()

	if !st.CheckSkill(eocontext.FilterSkillName) {
		t.Errorf("CheckSkill should return true for %s", eocontext.FilterSkillName)
	}
}

func TestBuildQuotaKeyAndTTL(t *testing.T) {
	ctx := newDummyHttpContext()
	ctx.SetLabel("product", "apinto")
	ctx.SetLabel("application", "app1")

	keyGen := context_label.NewKeyGenerator("{product}:quota-limiting:{strategy}:{target_type}:{application}:{period}:{time_format}")
	s := &Strategy{key: keyGen}

	now := time.Date(2026, 8, 13, 10, 30, 45, 0, time.UTC)

	tests := []struct {
		name            string
		period          quota_limiting_strategy.Period
		targetType      string
		expectedKeyPart string
		checkTTL        func(t *testing.T, ttl time.Duration)
	}{
		{
			name:            "Minute Period",
			period:          quota_limiting_strategy.PeriodMinute,
			targetType:      "user",
			expectedKeyPart: "apinto:quota-limiting:st1:user:app1:minute:202608131030",
			checkTTL: func(t *testing.T, ttl time.Duration) {
				expected := time.Duration(60-45)*time.Second + 10*time.Second
				if ttl != expected {
					t.Fatalf("minute ttl mismatch, expected %v, got %v", expected, ttl)
				}
			},
		},
		{
			name:            "Hour Period",
			period:          quota_limiting_strategy.PeriodHour,
			targetType:      "user",
			expectedKeyPart: "apinto:quota-limiting:st1:user:app1:hour:2026081310",
			checkTTL: func(t *testing.T, ttl time.Duration) {
				expected := time.Duration(3600-30*60-45)*time.Second + 60*time.Second
				if ttl != expected {
					t.Fatalf("hour ttl mismatch, expected %v, got %v", expected, ttl)
				}
			},
		},
		{
			name:            "Day Period",
			period:          quota_limiting_strategy.PeriodDay,
			targetType:      "channel",
			expectedKeyPart: "apinto:quota-limiting:st1:channel:app1:day:20260813",
			checkTTL: func(t *testing.T, ttl time.Duration) {
				expected := time.Duration(86400-10*3600-30*60-45)*time.Second + 300*time.Second
				if ttl != expected {
					t.Fatalf("day ttl mismatch, expected %v, got %v", expected, ttl)
				}
			},
		},
		{
			name:            "Month Period",
			period:          quota_limiting_strategy.PeriodMonth,
			targetType:      "user",
			expectedKeyPart: "apinto:quota-limiting:st1:user:app1:month:202608",
			checkTTL: func(t *testing.T, ttl time.Duration) {
				nextMonth := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
				expected := nextMonth.Sub(now) + 3600*time.Second
				if ttl != expected {
					t.Fatalf("month ttl mismatch, expected %v, got %v", expected, ttl)
				}
			},
		},
		{
			name:            "Total Period",
			period:          quota_limiting_strategy.PeriodTotal,
			targetType:      "user",
			expectedKeyPart: "apinto:quota-limiting:st1:user:app1:total:total",
			checkTTL: func(t *testing.T, ttl time.Duration) {
				if ttl != -1 {
					t.Fatalf("total ttl should be -1, got %v", ttl)
				}
			},
		},
		{
			name:            "Default Period",
			period:          quota_limiting_strategy.PeriodSecond,
			targetType:      "user",
			expectedKeyPart: "apinto:quota-limiting:st1:user:app1:second:20260813103045",
			checkTTL: func(t *testing.T, ttl time.Duration) {
				if ttl != 2*time.Second {
					t.Fatalf("default ttl should be 2s, got %v", ttl)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dummySt := &dummyStrategy{
				id:         "st1",
				targetType: tt.targetType,
				period:     tt.period,
			}
			key, ttl := s.buildQuotaKeyAndTTL(ctx, dummySt, now)
			if key != tt.expectedKeyPart {
				t.Fatalf("key mismatch:\nexpected: %s\ngot:      %s", tt.expectedKeyPart, key)
			}
			tt.checkTTL(t, ttl)
		})
	}
}

func TestDoHttpFilter(t *testing.T) {
	keyGen := context_label.NewKeyGenerator("{product}:quota-limiting:{strategy}:{target_type}:{application}:{period}:{time_format}")
	s := &Strategy{key: keyGen}

	t.Run("Scenario 1: No API label", func(t *testing.T) {
		ctx := newDummyHttpContext()
		chain := &dummyChain{}
		err := s.DoHttpFilter(ctx, chain)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if chain.called {
			t.Fatalf("chain should not be called when API label is empty")
		}
	})

	t.Run("Scenario 2: No strategies matched", func(t *testing.T) {
		ctx := newDummyHttpContext()
		ctx.SetLabel("api", "api_test")
		chain := &dummyChain{}

		err := s.DoHttpFilter(ctx, chain)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if !chain.called {
			t.Fatalf("chain should be called when no strategies matched")
		}
	})

	t.Run("Scenario 3: Strategy threshold <= 0", func(t *testing.T) {
		tenant := "tenant_zero"
		config := &quota_limiting_strategy.Config{
			Filters: quota_limiting_strategy.FiltersConfig{
				Tenant: tenant,
			},
			Quota: quota_limiting_strategy.QuotaConfig{
				Request: quota_limiting_strategy.QuotaRule{
					Second: 0, // <= 0
				},
			},
		}
		worker, err := quota_limiting_strategy.Create("strat_zero", "strat_zero", config, nil)
		if err != nil {
			t.Fatalf("create strategy error: %v", err)
		}
		defer worker.Stop()

		ctx := newDummyHttpContext()
		ctx.SetLabel("api", "api_test")
		ctx.SetLabel("tenant", tenant)
		chain := &dummyChain{}

		err = s.DoHttpFilter(ctx, chain)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if !chain.called {
			t.Fatalf("chain should be called when threshold <= 0")
		}
	})

	t.Run("Scenario 4: Request within quota (Allowed)", func(t *testing.T) {
		tenant := "tenant_allow"
		config := &quota_limiting_strategy.Config{
			Filters: quota_limiting_strategy.FiltersConfig{
				Tenant: tenant,
			},
			Quota: quota_limiting_strategy.QuotaConfig{
				Request: quota_limiting_strategy.QuotaRule{
					Minute: 5,
				},
			},
		}
		worker, err := quota_limiting_strategy.Create("strat_allow", "strat_allow", config, nil)
		if err != nil {
			t.Fatalf("create strategy error: %v", err)
		}
		defer worker.Stop()

		ctx := newDummyHttpContext()
		ctx.SetLabel("api", "api_test")
		ctx.SetLabel("tenant", tenant)
		ctx.SetLabel("product", "apinto")
		ctx.SetLabel("application", "app1")

		chain := &dummyChain{}

		err = s.DoHttpFilter(ctx, chain)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if !chain.called {
			t.Fatalf("chain should be executed when within quota")
		}
	})

	t.Run("Scenario 5: Request limit exceeded (Blocked & Default 429 Response)", func(t *testing.T) {
		tenant := "tenant_block"
		config := &quota_limiting_strategy.Config{
			Filters: quota_limiting_strategy.FiltersConfig{
				Tenant: tenant,
			},
			Quota: quota_limiting_strategy.QuotaConfig{
				Request: quota_limiting_strategy.QuotaRule{
					Minute: 1, // 阈值为 1
				},
			},
		}
		worker, err := quota_limiting_strategy.Create("strat_block", "strat_block", config, nil)
		if err != nil {
			t.Fatalf("create strategy error: %v", err)
		}
		defer worker.Stop()

		ctx := newDummyHttpContext()
		ctx.SetLabel("api", "api_test")
		ctx.SetLabel("tenant", tenant)
		ctx.SetLabel("product", "apinto")
		ctx.SetLabel("application", "app1")

		chain := &dummyChain{}

		// 第一次调用：递增到 1 (允许)
		err1 := s.DoHttpFilter(ctx, chain)
		if err1 != nil {
			t.Fatalf("first request should succeed, got %v", err1)
		}

		// 第二次调用：递增到 2，超过阈值 1 (应该拦截)
		chain2 := &dummyChain{}
		ctx2 := newDummyHttpContext()
		ctx2.SetLabel("api", "api_test")
		ctx2.SetLabel("tenant", tenant)
		ctx2.SetLabel("product", "apinto")
		ctx2.SetLabel("application", "app1")

		err2 := s.DoHttpFilter(ctx2, chain2)
		if !errors.Is(err2, ErrQuotaExceeded) {
			t.Fatalf("expected ErrQuotaExceeded, got %v", err2)
		}
		if chain2.called {
			t.Fatalf("chain should not be called when blocked")
		}

		// 验证上下文属性
		if isBlock, _ := ctx2.Value("is_block").(bool); !isBlock {
			t.Errorf("is_block should be true")
		}
		if handler := ctx2.GetLabel("handler"); handler != "quota-limiting-request" {
			t.Errorf("handler label expected 'quota-limiting-request', got '%s'", handler)
		}

		// 验证 HTTP 响应状态和 Body
		resp, ok := ctx2.Response().(*dummyHttpResponse)
		if !ok {
			t.Fatalf("expected *dummyHttpResponse")
		}
		if resp.StatusCode() != 429 {
			t.Errorf("expected http status code 429, got %d", resp.StatusCode())
		}
		expectedBody := `{"code":429,"message":"Request quota limit exceeded"}`
		if string(resp.GetBody()) != expectedBody {
			t.Errorf("expected body %s, got %s", expectedBody, string(resp.GetBody()))
		}
	})

	t.Run("Scenario 6: Custom Response on Blocked", func(t *testing.T) {
		customResp := &dummyCustomResponse{}
		st1 := &dummyStrategy{
			id:         "st_custom",
			targetType: "user",
			period:     quota_limiting_strategy.PeriodMinute,
			threshold:  1,
			resp:       customResp,
		}

		ctx := newDummyHttpContext()
		ctx.SetLabel("api", "api_test")
		ctx.SetLabel("product", "apinto")
		ctx.SetLabel("application", "app1")

		now := time.Now()
		key, ttl := s.buildQuotaKeyAndTTL(ctx, st1, now)

		cache := resources.LocalCache()
		cache.IncrBy(ctx.Context(), key, 1, ttl)

		val, _ := cache.IncrBy(ctx.Context(), key, 1, ttl).Result()
		if val > st1.Threshold() {
			st1.Response().Response(ctx)
		}

		if !customResp.called {
			t.Fatalf("custom response should be called")
		}
	})

	t.Run("Scenario 7: Cache IncrBy Error handling", func(t *testing.T) {
		mc := newMockCache()
		mc.incrErr = errors.New("redis connection reset")

		st1 := &dummyStrategy{
			id:        "st_err",
			period:    quota_limiting_strategy.PeriodMinute,
			threshold: 10,
		}

		ctx := newDummyHttpContext()
		now := time.Now()
		key, ttl := s.buildQuotaKeyAndTTL(ctx, st1, now)

		val, err := mc.IncrBy(ctx.Context(), key, 1, ttl).Result()
		if err == nil {
			t.Fatalf("expected error from mock cache")
		}
		if val != 0 {
			t.Fatalf("val should be 0 on error")
		}
	})
}

func TestDoFilter(t *testing.T) {
	keyGen := context_label.NewKeyGenerator("{product}:quota-limiting:{strategy}:{target_type}:{application}:{period}:{time_format}")
	s := &Strategy{key: keyGen}

	ctx := newDummyHttpContext()
	chain := &dummyChain{}

	err := s.DoFilter(ctx, chain)
	if err != nil {
		t.Fatalf("DoFilter returned error: %v", err)
	}
}

// ============================================================================
// Benchmark Tests
// ============================================================================

func BenchmarkBuildQuotaKeyAndTTL(b *testing.B) {
	keyGen := context_label.NewKeyGenerator("{product}:quota-limiting:{strategy}:{target_type}:{application}:{period}:{time_format}")
	s := &Strategy{key: keyGen}

	ctx := newDummyHttpContext()
	ctx.SetLabel("product", "apinto_benchmark")
	ctx.SetLabel("application", "app_bench_123")

	st := &dummyStrategy{
		id:         "strat_bench_999",
		targetType: "user",
		period:     quota_limiting_strategy.PeriodMinute,
		threshold:  10000,
	}

	now := time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = s.buildQuotaKeyAndTTL(ctx, st, now)
		}
	})
}

func BenchmarkDoHttpFilter_Allowed(b *testing.B) {
	tenant := "tenant_bench_allowed"
	config := &quota_limiting_strategy.Config{
		Filters: quota_limiting_strategy.FiltersConfig{
			Tenant: tenant,
		},
		Quota: quota_limiting_strategy.QuotaConfig{
			Request: quota_limiting_strategy.QuotaRule{
				Hour: 1000000000, // 足够大的阈值
			},
		},
	}
	worker, err := quota_limiting_strategy.Create("strat_bench_allow", "strat_bench_allow", config, nil)
	if err != nil {
		b.Fatalf("create strategy error: %v", err)
	}
	defer worker.Stop()

	keyGen := context_label.NewKeyGenerator("{product}:quota-limiting:{strategy}:{target_type}:{application}:{period}:{time_format}")
	s := &Strategy{key: keyGen}

	chain := &dummyChain{}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx := newDummyHttpContext()
		ctx.SetLabel("api", "api_bench")
		ctx.SetLabel("tenant", tenant)
		ctx.SetLabel("product", "apinto")
		ctx.SetLabel("application", "app1")

		_ = s.DoHttpFilter(ctx, chain)
	}
}

func BenchmarkDoHttpFilter_Blocked(b *testing.B) {
	tenant := "tenant_bench_blocked"
	config := &quota_limiting_strategy.Config{
		Filters: quota_limiting_strategy.FiltersConfig{
			Tenant: tenant,
		},
		Quota: quota_limiting_strategy.QuotaConfig{
			Request: quota_limiting_strategy.QuotaRule{
				Hour: 1, // 阈值为 1，几乎全被阻断
			},
		},
	}
	worker, err := quota_limiting_strategy.Create("strat_bench_block", "strat_bench_block", config, nil)
	if err != nil {
		b.Fatalf("create strategy error: %v", err)
	}
	defer worker.Stop()

	keyGen := context_label.NewKeyGenerator("{product}:quota-limiting:{strategy}:{target_type}:{application}:{period}:{time_format}")
	s := &Strategy{key: keyGen}

	// 先请求 1 次将其占满
	ctxInit := newDummyHttpContext()
	ctxInit.SetLabel("api", "api_bench")
	ctxInit.SetLabel("tenant", tenant)
	ctxInit.SetLabel("product", "apinto")
	ctxInit.SetLabel("application", "app1")
	chain := &dummyChain{}
	_ = s.DoHttpFilter(ctxInit, chain)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx := newDummyHttpContext()
		ctx.SetLabel("api", "api_bench")
		ctx.SetLabel("tenant", tenant)
		ctx.SetLabel("product", "apinto")
		ctx.SetLabel("application", "app1")

		_ = s.DoHttpFilter(ctx, chain)
	}
}

func BenchmarkDoHttpFilter_Parallel(b *testing.B) {
	tenant := "tenant_bench_parallel"
	config := &quota_limiting_strategy.Config{
		Filters: quota_limiting_strategy.FiltersConfig{
			Tenant: tenant,
		},
		Quota: quota_limiting_strategy.QuotaConfig{
			Request: quota_limiting_strategy.QuotaRule{
				Hour: 1000000000,
			},
		},
	}
	worker, err := quota_limiting_strategy.Create("strat_bench_parallel", "strat_bench_parallel", config, nil)
	if err != nil {
		b.Fatalf("create strategy error: %v", err)
	}
	defer worker.Stop()

	keyGen := context_label.NewKeyGenerator("{product}:quota-limiting:{strategy}:{target_type}:{application}:{period}:{time_format}")
	s := &Strategy{key: keyGen}

	chain := &dummyChain{}

	var seq int64

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			id := atomic.AddInt64(&seq, 1)
			ctx := newDummyHttpContext()
			ctx.SetLabel("api", "api_bench")
			ctx.SetLabel("tenant", tenant)
			ctx.SetLabel("product", "apinto")
			ctx.SetLabel("application", "app_"+string(rune(id%10)))

			_ = s.DoHttpFilter(ctx, chain)
		}
	})
}
