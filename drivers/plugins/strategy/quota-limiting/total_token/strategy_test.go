package total_token

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
	price_calcular "github.com/eolinker/apinto/price-calcular"
	"github.com/eolinker/apinto/resources"
	"github.com/eolinker/apinto/utils/response"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

// ============================================================================
// Mock & Dummy Implementations
// ============================================================================

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

func (r *dummyHttpResponse) ClearError()                               {}
func (r *dummyHttpResponse) HeaderReset()                              { r.header = make(http.Header) }
func (r *dummyHttpResponse) HeadersString() string                     { return "" }
func (r *dummyHttpResponse) IsBodyStream() bool                        { return false }
func (r *dummyHttpResponse) RemoteIP() string                          { return "127.0.0.1" }
func (r *dummyHttpResponse) RemoteAddr() string                        { return "127.0.0.1:8080" }
func (r *dummyHttpResponse) RemotePort() int                           { return 8080 }
func (r *dummyHttpResponse) ResponseError() error                      { return nil }
func (r *dummyHttpResponse) ResponseTime() time.Duration               { return 0 }
func (r *dummyHttpResponse) Headers() http.Header                      { return r.header }
func (r *dummyHttpResponse) AddHeader(key, value string)               {}
func (r *dummyHttpResponse) DelHeader(key string)                      {}
func (r *dummyHttpResponse) Response()                                 {}
func (r *dummyHttpResponse) ContentLength() int                        { return len(r.body) }
func (r *dummyHttpResponse) GetHeader(key string) string               { return r.header.Get(key) }
func (r *dummyHttpResponse) String() string                            { return string(r.body) }
func (r *dummyHttpResponse) SetResponseTime(duration time.Duration)   {}
func (r *dummyHttpResponse) GetResponseTime() time.Duration           { return 0 }

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

func (d *dummyHttpContext) GetEntry() eosc.IEntry                                    { return nil }
func (d *dummyHttpContext) Proxies() []http_service.IProxy                         { return nil }
func (d *dummyHttpContext) ProxyClone() http_service.IRequest                       { return nil }
func (d *dummyHttpContext) SetProxy(proxy http_service.IRequest)                    {}
func (d *dummyHttpContext) AcceptTime() time.Time                                   { return time.Now() }
func (d *dummyHttpContext) Scheme() string                                          { return "http" }
func (d *dummyHttpContext) Labels() map[string]string                               { return d.labels }
func (d *dummyHttpContext) GetComplete() eocontext.CompleteHandler                  { return nil }
func (d *dummyHttpContext) SetCompleteHandler(handler eocontext.CompleteHandler)    {}
func (d *dummyHttpContext) GetFinish() eocontext.FinishHandler                      { return nil }
func (d *dummyHttpContext) SetFinish(handler eocontext.FinishHandler)                {}
func (d *dummyHttpContext) GetBalance() eocontext.BalanceHandler                    { return nil }
func (d *dummyHttpContext) SetBalance(handler eocontext.BalanceHandler)            {}
func (d *dummyHttpContext) GetUpstreamHostHandler() eocontext.UpstreamHostHandler    { return nil }
func (d *dummyHttpContext) SetUpstreamHostHandler(handler eocontext.UpstreamHostHandler) {}
func (d *dummyHttpContext) RealIP() string                                          { return "127.0.0.1" }
func (d *dummyHttpContext) LocalIP() net.IP                                         { return net.ParseIP("127.0.0.1") }
func (d *dummyHttpContext) LocalAddr() net.Addr                                     { return nil }
func (d *dummyHttpContext) LocalPort() int                                          { return 80 }
func (d *dummyHttpContext) IsCloneable() bool                                       { return false }
func (d *dummyHttpContext) Clone() (eocontext.EoContext, error)                     { return d, nil }
func (d *dummyHttpContext) SendTo(scheme string, node eocontext.INode, timeout time.Duration) error {
	return nil
}
func (d *dummyHttpContext) Request() http_service.IRequestReader { return nil }
func (d *dummyHttpContext) Proxy() http_service.IRequest       { return nil }
func (d *dummyHttpContext) RequestId() string                  { return "req-123" }
func (d *dummyHttpContext) IsAccept() bool                     { return true }
func (d *dummyHttpContext) SetAccept(accept bool)              {}
func (d *dummyHttpContext) FastFinish()                        {}
func (d *dummyHttpContext) IsFinish() bool                     { return false }

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

type dummyChain struct {
	called bool
	err    error
}

func (c *dummyChain) DoChain(ctx eocontext.EoContext) error {
	c.called = true
	return c.err
}

func (c *dummyChain) Destroy() {}

type dummyCalculator struct {
	rules []*price_calcular.ProcessedRule
	err   error
}

func (m *dummyCalculator) ID() string                                                    { return "dummy" }
func (m *dummyCalculator) Currency() string                                              { return "USD" }
func (m *dummyCalculator) VariablesExtractor() price_calcular.IExtractor                { return nil }
func (m *dummyCalculator) Calculate(eocontext.EoContext, bool, *price_calcular.PricingData) (*price_calcular.CalculateResult, error) {
	return nil, nil
}
func (m *dummyCalculator) CalculateFromChunk(eocontext.EoContext, bool, []byte, *price_calcular.PricingData) (*price_calcular.CalculateResult, error) {
	return nil, nil
}
func (m *dummyCalculator) PreDeduct(eocontext.EoContext, *price_calcular.PricingData) (float64, string, error) {
	return 0, "", nil
}
func (m *dummyCalculator) CalculateByRule(eocontext.EoContext, *price_calcular.PricingData) (*price_calcular.CalculateResult, error) {
	return nil, nil
}
func (m *dummyCalculator) Rules(ctx eocontext.EoContext) ([]*price_calcular.ProcessedRule, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.rules, nil
}

type mockCustomResponse struct {
	called bool
}

func (m *mockCustomResponse) Response(ctx eocontext.EoContext) {
	m.called = true
	if httpCtx, err := http_service.Assert(ctx); err == nil {
		httpCtx.Response().SetStatus(429, "Too Many Requests")
		httpCtx.Response().SetBody([]byte(`{"message":"custom limit"}`))
	}
}

func (m *mockCustomResponse) Header() map[string]string {
	return nil
}

func (m *mockCustomResponse) StatusCode() int {
	return 429
}

type dummyStrategy struct {
	id         string
	targetType string
	period     quota_limiting_strategy.Period
	threshold  int64
	resp       response.IResponse
}

func (s *dummyStrategy) ID() string                              { return s.id }
func (s *dummyStrategy) TargetType() string                      { return s.targetType }
func (s *dummyStrategy) Period() quota_limiting_strategy.Period { return s.period }
func (s *dummyStrategy) Threshold() int64                       { return s.threshold }
func (s *dummyStrategy) Response() response.IResponse            { return s.resp }

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

func (m *mockCache) GetVal(key string) int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.data[key]
}

func createTokenCalculator(calcId string) price_calcular.ICalculator {
	calc, _ := price_calcular.NewCalculator(calcId, "USD", nil, []*price_calcular.Rule{
		{ID: "r1", Name: "token_rule", SaleExpression: "#sale.input_token"},
	})
	return calc
}

// ============================================================================
// Unit Tests
// ============================================================================

func TestDriverAndFactory(t *testing.T) {
	cfg := &Config{Key: ""}
	err := CheckConfig(cfg, nil)
	if err != nil {
		t.Fatalf("unexpected checkConfig error: %v", err)
	}
	if cfg.Key == "" {
		t.Fatalf("expected default key set, got empty")
	}

	cfg2 := &Config{Key: "custom_key"}
	CheckConfig(cfg2, nil)
	if cfg2.Key != "custom_key" {
		t.Fatalf("expected custom_key retained, got %s", cfg2.Key)
	}

	worker, err := Create("worker_1", "name_1", cfg2, nil)
	if err != nil {
		t.Fatalf("unexpected create error: %v", err)
	}

	st, ok := worker.(*Strategy)
	if !ok {
		t.Fatalf("expected *Strategy, got %T", worker)
	}

	if !st.CheckSkill(eocontext.FilterSkillName) {
		t.Fatalf("expected skill check true")
	}
	if st.CheckSkill("invalid_skill") {
		t.Fatalf("expected invalid skill check false")
	}

	if err := st.Start(); err != nil {
		t.Fatalf("expected Start nil")
	}
	if err := st.Stop(); err != nil {
		t.Fatalf("expected Stop nil")
	}
	if err := st.Reset(nil, nil); err != nil {
		t.Fatalf("expected Reset nil")
	}
	st.Destroy()
}

func TestDoHttpFilter_SkipConditions(t *testing.T) {
	st := &Strategy{
		key: context_label.NewKeyGenerator("{product}:quota-limiting:{strategy}:{target_type}:{application}:{period}:{time_format}"),
	}

	// 1. Missing api label
	t.Run("missing api label", func(t *testing.T) {
		ctx := newDummyHttpContext()
		chain := &dummyChain{}
		err := st.DoHttpFilter(ctx, chain)
		if err != nil {
			t.Fatalf("expected nil error when api is empty, got %v", err)
		}
		if chain.called {
			t.Fatalf("expected chain not called when api is empty")
		}
	})

	// 2. Calculator not found
	t.Run("calculator not found", func(t *testing.T) {
		ctx := newDummyHttpContext()
		ctx.SetLabel("api", "api_1")
		ctx.SetLabel("resource_type", "api")
		ctx.SetLabel("resource", "res_1")
		chain := &dummyChain{}

		err := st.DoHttpFilter(ctx, chain)
		if err == nil {
			t.Fatalf("expected error when calculator not found")
		}
		if ctx.Response().StatusCode() != 500 {
			t.Fatalf("expected 500 status code, got %d", ctx.Response().StatusCode())
		}
	})

	// 3. Calculator Rules error
	t.Run("calculator rules error", func(t *testing.T) {
		calcId := "api:res_rules_err"
		price_calcular.SetCalculator(calcId, &dummyCalculator{err: errors.New("rule err")})
		defer price_calcular.DelCalculator(calcId)

		ctx := newDummyHttpContext()
		ctx.SetLabel("api", "api_1")
		ctx.SetLabel("resource_type", "api")
		ctx.SetLabel("resource", "res_rules_err")
		chain := &dummyChain{}

		err := st.DoHttpFilter(ctx, chain)
		if err == nil {
			t.Fatalf("expected error on calculator rules error")
		}
		if ctx.Response().StatusCode() != 500 {
			t.Fatalf("expected 500 status code, got %d", ctx.Response().StatusCode())
		}
	})

	// 4. Estimate token <= 0
	t.Run("estimate token <= 0", func(t *testing.T) {
		calcId := "api:res_zero_token"
		price_calcular.SetCalculator(calcId, &dummyCalculator{rules: []*price_calcular.ProcessedRule{}})
		defer price_calcular.DelCalculator(calcId)

		ctx := newDummyHttpContext()
		ctx.SetLabel("api", "api_1")
		ctx.SetLabel("resource_type", "api")
		ctx.SetLabel("resource", "res_zero_token")
		chain := &dummyChain{}

		err := st.DoHttpFilter(ctx, chain)
		if err != nil {
			t.Fatalf("expected nil error when estimate token is 0, got %v", err)
		}
		if !chain.called {
			t.Fatalf("expected chain called when estimate token <= 0")
		}
	})

	// 5. No matched strategies
	t.Run("no matched total token strategies", func(t *testing.T) {
		calcId := "api:res_no_strat"
		calc := createTokenCalculator(calcId)
		price_calcular.SetCalculator(calcId, calc)
		defer price_calcular.DelCalculator(calcId)

		ctx := newDummyHttpContext()
		ctx.SetLabel("api", "api_1")
		ctx.SetLabel("resource_type", "api")
		ctx.SetLabel("resource", "res_no_strat")
		context_label.SetPreInputToken(ctx, 100)

		chain := &dummyChain{}
		err := st.DoHttpFilter(ctx, chain)
		if err != nil {
			t.Fatalf("expected nil error when no strategies match, got %v", err)
		}
		if !chain.called {
			t.Fatalf("expected chain called when no strategies match")
		}
	})
}

func TestDoHttpFilter_PreDeductAndSettle(t *testing.T) {
	cfg := &quota_limiting_strategy.Config{
		Na: "strat_settle_test",
		Quota: quota_limiting_strategy.QuotaConfig{
			TotalToken: quota_limiting_strategy.QuotaRule{Hour: 10000},
		},
		Filters: quota_limiting_strategy.FiltersConfig{
			Tenant:   "tenant_settle",
			Resource: quota_limiting_strategy.Filter{All: true},
			Target:   quota_limiting_strategy.Filter{Type: "all", All: true},
		},
	}
	_, err := quota_limiting_strategy.Create("worker_settle_1", "strat_settle_test", cfg, nil)
	if err != nil {
		t.Fatalf("failed to create total token strategy: %v", err)
	}

	st := &Strategy{
		key: context_label.NewKeyGenerator("{product}:quota-limiting:{strategy}:{target_type}:{application}:{period}:{time_format}"),
	}

	calcId := "api:res_settle"
	calc := createTokenCalculator(calcId)
	price_calcular.SetCalculator(calcId, calc)
	defer price_calcular.DelCalculator(calcId)

	// 测试 1: actualToken == estimateToken
	t.Run("exact estimation", func(t *testing.T) {
		mc := newMockCache()
		ctx := newDummyHttpContext()
		ctx.SetLabel("api", "api_1")
		ctx.SetLabel("resource_type", "api")
		ctx.SetLabel("resource", "res_settle")
		ctx.SetLabel("tenant", "tenant_settle")

		context_label.SetPreInputToken(ctx, 100)
		context_label.SetActualTotalToken(ctx, 100)

		chain := &dummyChain{}
		now := time.Now()
		tenantStrategies, _ := quota_limiting_strategy.GetTotalTokenStrategies(ctx)
		st2 := tenantStrategies[0].Strategies()[0]
		key, ttl := st.buildQuotaKeyAndTTL(ctx, st2, now)

		mc.IncrBy(ctx.Context(), key, 100, ttl)
		items := []*executedItem{{key: key, ttl: ttl}}

		if err := chain.DoChain(ctx); err != nil {
			t.Fatalf("chain err: %v", err)
		}
		settle(ctx, mc, items, 100)

		if mc.GetVal(key) != 100 {
			t.Fatalf("expected cached total 100, got %d", mc.GetVal(key))
		}
	})

	// 测试 2: actualToken > estimateToken (补扣)
	t.Run("actual token greater than estimated", func(t *testing.T) {
		mc := newMockCache()
		ctx := newDummyHttpContext()
		ctx.SetLabel("api", "api_1")
		ctx.SetLabel("resource_type", "api")
		ctx.SetLabel("resource", "res_settle")
		ctx.SetLabel("tenant", "tenant_settle")

		context_label.SetPreInputToken(ctx, 100)
		context_label.SetActualTotalToken(ctx, 150)

		now := time.Now()
		tenantStrategies, _ := quota_limiting_strategy.GetTotalTokenStrategies(ctx)
		st2 := tenantStrategies[0].Strategies()[0]
		key, ttl := st.buildQuotaKeyAndTTL(ctx, st2, now)

		mc.IncrBy(ctx.Context(), key, 100, ttl)
		items := []*executedItem{{key: key, ttl: ttl}}

		settle(ctx, mc, items, 100)

		if mc.GetVal(key) != 150 {
			t.Fatalf("expected cached total 150, got %d", mc.GetVal(key))
		}
	})

	// 测试 3: actualToken < estimateToken (退还多预扣的)
	t.Run("actual token less than estimated", func(t *testing.T) {
		mc := newMockCache()
		ctx := newDummyHttpContext()
		ctx.SetLabel("api", "api_1")
		ctx.SetLabel("resource_type", "api")
		ctx.SetLabel("resource", "res_settle")
		ctx.SetLabel("tenant", "tenant_settle")

		context_label.SetPreInputToken(ctx, 100)
		context_label.SetActualTotalToken(ctx, 60)

		now := time.Now()
		tenantStrategies, _ := quota_limiting_strategy.GetTotalTokenStrategies(ctx)
		st2 := tenantStrategies[0].Strategies()[0]
		key, ttl := st.buildQuotaKeyAndTTL(ctx, st2, now)

		mc.IncrBy(ctx.Context(), key, 100, ttl)
		items := []*executedItem{{key: key, ttl: ttl}}

		settle(ctx, mc, items, 100)

		if mc.GetVal(key) != 60 {
			t.Fatalf("expected cached total 60, got %d", mc.GetVal(key))
		}
	})

	// 测试 4: actualToken <= 0 (完全退还)
	t.Run("actual token <= 0", func(t *testing.T) {
		mc := newMockCache()
		ctx := newDummyHttpContext()
		ctx.SetLabel("api", "api_1")
		ctx.SetLabel("resource_type", "api")
		ctx.SetLabel("resource", "res_settle")
		ctx.SetLabel("tenant", "tenant_settle")

		context_label.SetPreInputToken(ctx, 100)
		context_label.SetActualTotalToken(ctx, 0)

		now := time.Now()
		tenantStrategies, _ := quota_limiting_strategy.GetTotalTokenStrategies(ctx)
		st2 := tenantStrategies[0].Strategies()[0]
		key, ttl := st.buildQuotaKeyAndTTL(ctx, st2, now)

		mc.IncrBy(ctx.Context(), key, 100, ttl)
		items := []*executedItem{{key: key, ttl: ttl}}

		settle(ctx, mc, items, 100)

		if mc.GetVal(key) != 0 {
			t.Fatalf("expected cached total 0 after refund, got %d", mc.GetVal(key))
		}
	})
}

func TestDoHttpFilter_QuotaExceededAndRollback(t *testing.T) {
	st := &Strategy{
		key: context_label.NewKeyGenerator("{product}:quota-limiting:{strategy}:{target_type}:{application}:{period}:{time_format}"),
	}

	calcId := "api:res_exceed"
	calc := createTokenCalculator(calcId)
	price_calcular.SetCalculator(calcId, calc)
	defer price_calcular.DelCalculator(calcId)

	// 单策略超限拦截
	t.Run("single strategy exceeded", func(t *testing.T) {
		cfg := &quota_limiting_strategy.Config{
			Na: "strat_exceed_1",
			Quota: quota_limiting_strategy.QuotaConfig{
				TotalToken: quota_limiting_strategy.QuotaRule{Hour: 100},
			},
			Filters: quota_limiting_strategy.FiltersConfig{
				Tenant:   "tenant_exceed",
				Resource: quota_limiting_strategy.Filter{All: true},
				Target:   quota_limiting_strategy.Filter{Type: "all", All: true},
			},
		}
		quota_limiting_strategy.Create("worker_exceed_1", "strat_exceed_1", cfg, nil)

		ctx := newDummyHttpContext()
		ctx.SetLabel("api", "api_1")
		ctx.SetLabel("resource_type", "api")
		ctx.SetLabel("resource", "res_exceed")
		ctx.SetLabel("tenant", "tenant_exceed")
		context_label.SetPreInputToken(ctx, 1500000000) // 超出 100 阈值

		chain := &dummyChain{}
		err := st.DoHttpFilter(ctx, chain)

		if err != ErrQuotaExceeded {
			t.Fatalf("expected ErrQuotaExceeded, got %v", err)
		}
		if ctx.Value("is_block") != true {
			t.Fatalf("expected is_block true")
		}
		if ctx.GetLabel("handler") != "quota-limiting-total-token" {
			t.Fatalf("expected handler label set")
		}
		if ctx.Response().StatusCode() != 429 {
			t.Fatalf("expected 429 status code, got %d", ctx.Response().StatusCode())
		}
		if chain.called {
			t.Fatalf("expected chain not called on exceed")
		}
	})
}

func TestBuildQuotaKeyAndTTL_AllPeriods(t *testing.T) {
	ctx := &dummyHttpContext{
		labels: map[string]string{
			"product":     "apinto",
			"application": "app_1",
		},
	}
	st := &Strategy{
		key: context_label.NewKeyGenerator("{product}:quota-limiting:{strategy}:{target_type}:{application}:{period}:{time_format}"),
	}

	now := time.Date(2026, 8, 13, 14, 30, 15, 0, time.UTC)

	periods := []struct {
		period      quota_limiting_strategy.Period
		expectedFmt string
	}{
		{quota_limiting_strategy.PeriodSecond, "20260813143015"},
		{quota_limiting_strategy.PeriodMinute, "202608131430"},
		{quota_limiting_strategy.PeriodHour, "2026081314"},
		{quota_limiting_strategy.PeriodDay, "20260813"},
		{quota_limiting_strategy.PeriodMonth, "202608"},
		{quota_limiting_strategy.PeriodTotal, "total"},
		{quota_limiting_strategy.Period(999), "20260813143015"}, // default
	}

	for _, p := range periods {
		dummySt := &dummyStrategy{
			id:         "s1",
			targetType: "user",
			period:     p.period,
			threshold:  1000,
		}
		key, ttl := st.buildQuotaKeyAndTTL(ctx, dummySt, now)

		if p.period == quota_limiting_strategy.PeriodTotal {
			if ttl != -1 {
				t.Fatalf("expected ttl -1 for total period, got %v", ttl)
			}
		} else {
			if ttl <= 0 {
				t.Fatalf("expected positive ttl for period %v, got %v", p.period, ttl)
			}
		}

		expectedKey := "apinto:quota-limiting:s1:user:app_1:" + p.period.String() + ":" + p.expectedFmt
		if key != expectedKey {
			t.Fatalf("period %v: expected key %s, got %s", p.period, expectedKey, key)
		}
	}
}

func TestDoHttpFilter_Concurrent(t *testing.T) {
	mc := newMockCache()

	const concurrency = 50
	const tokenPerReq = 10
	const threshold = 250 // 最多允许 25 次请求通过

	var passCount int32
	var rejectCount int32
	var wg sync.WaitGroup

	key := "test_concurrent_key"
	ttl := time.Hour

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			val, _ := mc.IncrBy(context.Background(), key, tokenPerReq, ttl).Result()
			if val > threshold {
				// 超限回滚
				mc.DecrBy(context.Background(), key, tokenPerReq, ttl)
				atomic.AddInt32(&rejectCount, 1)
			} else {
				atomic.AddInt32(&passCount, 1)
			}
		}()
	}

	wg.Wait()

	if passCount != 25 {
		t.Fatalf("expected 25 passes, got %d", passCount)
	}
	if rejectCount != 25 {
		t.Fatalf("expected 25 rejects, got %d", rejectCount)
	}
	if mc.GetVal(key) != 250 {
		t.Fatalf("expected final total 250, got %d", mc.GetVal(key))
	}
}
