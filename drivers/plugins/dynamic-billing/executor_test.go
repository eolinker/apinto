package dynamic_billing

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	context_label "github.com/eolinker/apinto/common/context-label"
	"github.com/eolinker/apinto/drivers"
	price_calcular "github.com/eolinker/apinto/price-calcular"
	"github.com/eolinker/apinto/resources"
	scope_manager "github.com/eolinker/apinto/scope-manager"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
)

// ============================================================================
// Cache mocks
// ============================================================================

// mockStringResult 模拟 ICache.Get 的返回结果
type mockStringResult struct {
	val string
	err error
}

func (m *mockStringResult) Result() (string, error) { return m.val, m.err }
func (m *mockStringResult) Bytes() ([]byte, error)  { return []byte(m.val), m.err }

// mockIntResult 模拟 IntResult 返回结果
type mockIntResult struct {
	val int64
	err error
}

func (m *mockIntResult) Result() (int64, error) { return m.val, m.err }

// mockInterfaceResult 模拟 InterfaceResult 返回结果
type mockInterfaceResult struct {
	val interface{}
	err error
}

func (m *mockInterfaceResult) Result() (interface{}, error) { return m.val, m.err }

// mockBoolResult 模拟 BoolResult 返回结果
type mockBoolResult struct {
	val bool
	err error
}

func (m *mockBoolResult) Result() (bool, error) { return m.val, m.err }

// mockCache 模拟 Redis 缓存资源，kv 作为内存存储，支持并发计数与 Lua 扣减记录
type mockCache struct {
	resources.ICache
	val string
	err error
	kv  map[string]string
	// runCalls 记录每次 Run(Lua) 的参数，便于断言扣款调用
	runCalls []runCall
}

type runCall struct {
	keys []string
	args []interface{}
}

func (m *mockCache) ensureKV() {
	if m.kv == nil {
		m.kv = make(map[string]string)
	}
}

func (m *mockCache) Get(ctx context.Context, key string) resources.StringResult {
	if m.kv != nil {
		if v, ok := m.kv[key]; ok {
			return &mockStringResult{val: v, err: nil}
		}
	}
	return &mockStringResult{val: m.val, err: m.err}
}

func (m *mockCache) Set(ctx context.Context, key string, value []byte, expiration time.Duration) resources.StatusResult {
	m.ensureKV()
	m.kv[key] = string(value)
	return nil
}

func (m *mockCache) SetNX(ctx context.Context, key string, value []byte, expiration time.Duration) resources.BoolResult {
	m.ensureKV()
	if _, ok := m.kv[key]; ok {
		return &mockBoolResult{val: false, err: nil}
	}
	m.kv[key] = string(value)
	return &mockBoolResult{val: true, err: nil}
}

func (m *mockCache) IncrBy(ctx context.Context, key string, decrement int64, expiration time.Duration) resources.IntResult {
	return &mockIntResult{val: 1, err: nil}
}

func (m *mockCache) DecrBy(ctx context.Context, key string, decrement int64, expiration time.Duration) resources.IntResult {
	return &mockIntResult{val: 0, err: nil}
}

func (m *mockCache) Run(ctx context.Context, script interface{}, keys []string, args ...interface{}) resources.InterfaceResult {
	m.runCalls = append(m.runCalls, runCall{keys: keys, args: args})
	return &mockInterfaceResult{val: int64(1), err: nil}
}

func (m *mockCache) Id() string                                               { return "mock_cache" }
func (m *mockCache) Name() string                                             { return "mock_cache" }
func (m *mockCache) Start() error                                             { return nil }
func (m *mockCache) Reset(interface{}, map[eosc.RequireId]eosc.IWorker) error { return nil }
func (m *mockCache) Stop() error                                              { return nil }
func (m *mockCache) CheckSkill(string) bool                                   { return true }

// ============================================================================
// HTTP context mocks
// ============================================================================

// mockResponse 模拟 HTTP 响应
type mockResponse struct {
	http_context.IResponse
	statusCode int
	body       []byte
	headers    http.Header
	isStream   bool
}

func (m *mockResponse) StatusCode() int         { return m.statusCode }
func (m *mockResponse) ContentLength() int      { return len(m.body) }
func (m *mockResponse) ContentEncoding() []byte { return nil }
func (m *mockResponse) GetBody() []byte         { return m.body }
func (m *mockResponse) SetBody(b []byte)        { m.body = b }
func (m *mockResponse) IsBodyStream() bool      { return m.isStream }
func (m *mockResponse) SetStatus(code int, status string) {
	m.statusCode = code
}
func (m *mockResponse) SetHeader(key, value string) {
	if m.headers == nil {
		m.headers = make(http.Header)
	}
	m.headers.Set(key, value)
}
func (m *mockResponse) Headers() http.Header {
	if m.headers == nil {
		m.headers = make(http.Header)
	}
	return m.headers
}
func (m *mockResponse) ResponseTime() time.Duration { return 0 }
func (m *mockResponse) ResponseError() error        { return nil }

// mockProxy 模拟转发请求对象，收集注册的回调以便测试驱动执行
type mockProxy struct {
	http_context.IRequest
	bodyFinishFns  []http_context.BodyFinishFunc
	streamBodyFns  []http_context.StreamFunc
	streamBodyFunc http_context.StreamParseFunc
}

func (m *mockProxy) GetStreamBodyParse() http_context.StreamParseFunc { return m.streamBodyFunc }
func (m *mockProxy) AppendBodyFinish(fn http_context.BodyFinishFunc) {
	m.bodyFinishFns = append(m.bodyFinishFns, fn)
}
func (m *mockProxy) AppendStreamBodyHandle(fn http_context.StreamFunc) {
	m.streamBodyFns = append(m.streamBodyFns, fn)
}

// mockHttpContext 模拟 http_context.IHttpContext
type mockHttpContext struct {
	http_context.IHttpContext
	labels map[string]string
	values map[string]interface{}
	ctx    context.Context
	resp   *mockResponse
	proxy  *mockProxy
}

func newMockHttpContext(labels map[string]string, resp *mockResponse) *mockHttpContext {
	return &mockHttpContext{
		labels: labels,
		values: make(map[string]interface{}),
		ctx:    context.Background(),
		resp:   resp,
		proxy:  &mockProxy{},
	}
}

func (m *mockHttpContext) GetLabel(key string) string { return m.labels[key] }
func (m *mockHttpContext) SetLabel(key, val string)   { m.labels[key] = val }

func (m *mockHttpContext) Assert(i interface{}) error {
	if v, ok := i.(*http_context.IHttpContext); ok {
		*v = m
		return nil
	}
	return nil
}

func (m *mockHttpContext) WithValue(key interface{}, val interface{}) {
	if k, ok := key.(string); ok {
		m.values[k] = val
	}
}

func (m *mockHttpContext) Value(key interface{}) interface{} {
	if k, ok := key.(string); ok {
		return m.values[k]
	}
	return nil
}

func (m *mockHttpContext) Context() context.Context             { return m.ctx }
func (m *mockHttpContext) Response() http_context.IResponse     { return m.resp }
func (m *mockHttpContext) Request() http_context.IRequestReader { return nil }
func (m *mockHttpContext) Proxy() http_context.IRequest         { return m.proxy }

// mockChain 模拟调用链
type mockChain struct {
	eocontext.IChain
}

func (m *mockChain) DoChain(ctx eocontext.EoContext) error { return nil }

// ============================================================================
// helpers
// ============================================================================

// newTestCalculator 构造一条命中 status==200 的成功计费规则计算器。
func newTestCalculator(t *testing.T) price_calcular.ICalculator {
	t.Helper()
	variables := price_calcular.Variables{
		"status": {
			Source: "response_status",
			Type:   "integer",
		},
	}
	rules := []*price_calcular.Rule{
		{
			ID:   "rule_success",
			Name: "成功请求规则",
			Conditions: &price_calcular.Condition{
				AllOf: []*price_calcular.BasicRule{
					{
						Key:   "status",
						Op:    "==",
						Value: "200",
						Type:  "integer",
					},
				},
			},
			CostExpression:     "cost_per_call * 1.0",
			SaleExpression:     "sale_per_call * 1.1",
			OfficialExpression: "official_per_call * 1.2",
		},
	}
	calc, err := price_calcular.NewCalculator("USD", variables, rules)
	if err != nil {
		t.Fatalf("create calculator failed: %v", err)
	}
	return calc
}

// ============================================================================
// tests
// ============================================================================

// TestDynamicBilling_ImmediateSettle 验证同步（immediate）模式下命中规则并完成计费与扣款。
func TestDynamicBilling_ImmediateSettle(t *testing.T) {
	// 注册资源计算器，key 为 resource_type:resource
	price_calcular.SetCalculator("api:res_01", newTestCalculator(t))
	defer price_calcular.DelCalculator("api:res_01")

	// mock Redis 定价数据
	redisPriceJSON := `{
		"basic_info": {"version": "v1.0.0", "resource_group_id": "res_01", "tenant_id": "tenant_01"},
		"strategy": {
			"rule_success": {
				"cost": {"per_call": 2.0},
				"sale": {"per_call": 20.0},
				"official": {"per_call": 25.0}
			}
		}
	}`
	cache := &mockCache{val: redisPriceJSON}
	scope_manager.Set("mock_cache", cache, "redis")
	defer scope_manager.Del("mock_cache")

	plugin := &executor{
		WorkerBase:              drivers.Worker("pricing_filter_test", "pricing_filter_test"),
		redisID:                 "mock_cache",
		enableBalance:           true,
		balanceKeyGenerator:     context_label.NewKeyGenerator("balance:{application}"),
		priceKeyGenerator:       context_label.NewKeyGenerator("access-resource-price:{application}:{resource}"),
		taskKeyGenerator:        context_label.NewKeyGenerator("dynamic-billing-task:{application}:{resource}"),
		concurrencyKeyGenerator: context_label.NewKeyGenerator("dynamic-billing-concurrency:{application}:{resource}"),
	}

	ctx := newMockHttpContext(map[string]string{
		"application":   "app_client_01",
		"api":           "api_01",
		"resource":      "res_01",
		"resource_type": "api",
	}, &mockResponse{statusCode: 200})

	if err := plugin.DoHttpFilter(ctx, &mockChain{}); err != nil {
		t.Fatalf("DoHttpFilter failed: %v", err)
	}

	// cost_per_call*1.0 = 2.0；sale_per_call*1.1 = 22.0；official 表达式实际复用 sale 表达式 => 22.0
	if got := context_label.GetAmountCost(ctx); got != "2.000000" {
		t.Errorf("expected cost_amount 2.000000, got %q", got)
	}
	if got := context_label.GetAmountSale(ctx); got != "22.000000" {
		t.Errorf("expected sale_amount 22.000000, got %q", got)
	}

	// sale > 0，应触发一次余额扣减 Lua 调用
	if len(cache.runCalls) != 1 {
		t.Fatalf("expected 1 balance deduction call, got %d", len(cache.runCalls))
	}
	// 扣款金额按 6 位放大为整数：22.0 * 1000000 = 22000000
	if got := cache.runCalls[0].args[0]; got != int64(22000000) {
		t.Errorf("expected deduct arg 22000000, got %v", got)
	}
}

// TestDynamicBilling_InsufficientBalance 验证余额不足时前置阻断返回 402。
func TestDynamicBilling_InsufficientBalance(t *testing.T) {
	price_calcular.SetCalculator("api:res_01", newTestCalculator(t))
	defer price_calcular.DelCalculator("api:res_01")

	// 需要同时提供 balance 与 price 两个 key：balance=0 触发余额不足，price 用于反序列化定价数据
	priceJSON := `{
		"basic_info": {"version": "v1.0.0", "resource_group_id": "res_01", "tenant_id": "tenant_01"},
		"strategy": {
			"rule_success": {
				"cost": {"per_call": 2.0},
				"sale": {"per_call": 20.0},
				"official": {"per_call": 25.0}
			}
		}
	}`
	cache := &mockCache{kv: map[string]string{
		"balance:app_client_01":                      "0",
		"access-resource-price:app_client_01:res_01": priceJSON,
	}}
	scope_manager.Set("mock_cache", cache, "redis")
	defer scope_manager.Del("mock_cache")

	plugin := &executor{
		WorkerBase:              drivers.Worker("pricing_filter_test", "pricing_filter_test"),
		redisID:                 "mock_cache",
		enableBalance:           true,
		balanceKeyGenerator:     context_label.NewKeyGenerator("balance:{application}"),
		priceKeyGenerator:       context_label.NewKeyGenerator("access-resource-price:{application}:{resource}"),
		taskKeyGenerator:        context_label.NewKeyGenerator("dynamic-billing-task:{application}:{resource}"),
		concurrencyKeyGenerator: context_label.NewKeyGenerator("dynamic-billing-concurrency:{application}:{resource}"),
	}

	resp := &mockResponse{statusCode: 200}
	ctx := newMockHttpContext(map[string]string{
		"application":   "app_client_01",
		"api":           "api_01",
		"resource":      "res_01",
		"resource_type": "api",
	}, resp)

	err := plugin.DoHttpFilter(ctx, &mockChain{})
	if err == nil {
		t.Fatalf("expected insufficient balance error, got nil")
	}
	if resp.statusCode != http.StatusPaymentRequired {
		t.Errorf("expected status 402, got %d", resp.statusCode)
	}
	// 余额不足时不应发生扣款
	if len(cache.runCalls) != 0 {
		t.Errorf("expected no deduction call, got %d", len(cache.runCalls))
	}
}

// TestDynamicBilling_TaskPricingSnapshot 验证异步两阶段：create 阶段固化定价快照，
// query 阶段即使 Redis 实时价格被篡改，仍按快照价格结算。
func TestDynamicBilling_TaskPricingSnapshot(t *testing.T) {
	price_calcular.SetCalculator("api:res_snapshot", newTestCalculator(t))
	defer price_calcular.DelCalculator("api:res_snapshot")

	redisPriceJSON := `{
		"basic_info": {"version": "v1.0.0", "resource_group_id": "res_snapshot", "tenant_id": "tenant_01"},
		"strategy": {
			"rule_success": {
				"cost": {"per_call": 10.0},
				"sale": {"per_call": 10.0},
				"official": {"per_call": 10.0}
			}
		}
	}`
	cache := &mockCache{val: redisPriceJSON, kv: make(map[string]string)}

	priceKey := "access-resource-price:user_snapshot:res_snapshot"
	cache.kv[priceKey] = redisPriceJSON

	scope_manager.Set("mock_cache_snapshot", cache, "redis")
	defer scope_manager.Del("mock_cache_snapshot")

	plugin := &executor{
		WorkerBase:              drivers.Worker("pricing_filter_test", "pricing_filter_test"),
		redisID:                 "mock_cache_snapshot",
		defaultConcurrencyLimit: 100,
		enableBalance:           true,
		balanceKeyGenerator:     context_label.NewKeyGenerator("balance:{application}"),
		priceKeyGenerator:       context_label.NewKeyGenerator("access-resource-price:{application}:{resource}"),
		taskKeyGenerator:        context_label.NewKeyGenerator("dynamic-billing-task:{application}:{resource}"),
		concurrencyKeyGenerator: context_label.NewKeyGenerator("dynamic-billing-concurrency:{application}:{resource}"),
	}

	// ---- create 阶段：应固化定价快照到 Redis ----
	ctxCreate := newMockHttpContext(map[string]string{
		"application":   "user_snapshot",
		"api":           "api_task",
		"resource":      "res_snapshot",
		"resource_type": "api",
		"billing_mode":  context_label.BillingModeTaskCreate,
	}, &mockResponse{statusCode: 200})

	if err := plugin.DoHttpFilter(ctxCreate, &mockChain{}); err != nil {
		t.Fatalf("TaskCreate DoHttpFilter failed: %v", err)
	}

	taskKey := "dynamic-billing-task:user_snapshot:res_snapshot"
	taskInfoStr, hasTask := cache.kv[taskKey]
	if !hasTask {
		t.Fatalf("expected TaskInfo in redis, but not found")
	}

	var taskInfo TaskInfo
	if err := json.Unmarshal([]byte(taskInfoStr), &taskInfo); err != nil {
		t.Fatalf("unmarshal TaskInfo failed: %v", err)
	}
	if taskInfo.PricingData == nil {
		t.Fatalf("expected pricing data snapshotted in TaskInfo, but got nil")
	}
	if plan := taskInfo.PricingData.Strategy["rule_success"]; plan == nil || plan.Sale["per_call"] != 10.0 {
		t.Fatalf("expected snapshotted sale per_call 10.0, got %+v", taskInfo.PricingData.Strategy["rule_success"])
	}

	// ---- 篡改 Redis 实时价格为 50.0 ----
	cache.kv[priceKey] = `{
		"basic_info": {"version": "v1.0.0", "resource_group_id": "res_snapshot", "tenant_id": "tenant_01"},
		"strategy": {
			"rule_success": {
				"cost": {"per_call": 50.0},
				"sale": {"per_call": 50.0},
				"official": {"per_call": 50.0}
			}
		}
	}`

	// ---- query 阶段：应读取快照数据，快照 sale per_call 仍为 10.0，而非被篡改的 50.0 ----
	ctxQuery := newMockHttpContext(map[string]string{
		"application":   "user_snapshot",
		"api":           "api_task",
		"resource":      "res_snapshot",
		"resource_type": "api",
		"billing_mode":  context_label.BillingModeTaskQuery,
	}, &mockResponse{statusCode: 200})

	if err := plugin.DoHttpFilter(ctxQuery, &mockChain{}); err != nil {
		t.Fatalf("TaskQuery DoHttpFilter failed: %v", err)
	}
}

// ============================================================================
// 纯函数/配置/工厂相关单测
// ============================================================================

// TestCheckConfig_Defaults 校验缺省值补齐逻辑
func TestCheckConfig_Defaults(t *testing.T) {
	cfg := &Config{}
	if err := checkConfig(cfg, nil); err != nil {
		t.Fatalf("checkConfig error: %v", err)
	}
	if cfg.BalanceKey != "balance:{balance_target}" {
		t.Errorf("BalanceKey default wrong: %q", cfg.BalanceKey)
	}
	if cfg.PriceKey != "access-resource-price:{application}:{resource}" {
		t.Errorf("PriceKey default wrong: %q", cfg.PriceKey)
	}
	if cfg.TaskKey != "resource-pricing-task:{application}:{resource}" {
		t.Errorf("TaskKey default wrong: %q", cfg.TaskKey)
	}
	if cfg.ConcurrencyKey != "resource-pricing-concurrency:{application}:{resource}" {
		t.Errorf("ConcurrencyKey default wrong: %q", cfg.ConcurrencyKey)
	}
}

// TestCheckConfig_KeepsCustom 已配置的字段不应被默认值覆盖
func TestCheckConfig_KeepsCustom(t *testing.T) {
	cfg := &Config{
		BalanceKey:     "custom-balance:{application}",
		PriceKey:       "custom-price:{resource}",
		TaskKey:        "custom-task:{application}",
		ConcurrencyKey: "custom-concurrency:{resource}",
	}
	if err := checkConfig(cfg, nil); err != nil {
		t.Fatalf("checkConfig error: %v", err)
	}
	if cfg.BalanceKey != "custom-balance:{application}" ||
		cfg.PriceKey != "custom-price:{resource}" ||
		cfg.TaskKey != "custom-task:{application}" ||
		cfg.ConcurrencyKey != "custom-concurrency:{resource}" {
		t.Errorf("checkConfig unexpectedly overwrote custom values: %+v", cfg)
	}
}

// TestHasFreePricePlan 覆盖免费策略识别的各种边界
func TestHasFreePricePlan(t *testing.T) {
	cases := []struct {
		name string
		data *price_calcular.PricingData
		want bool
	}{
		{name: "nil pricing data", data: nil, want: false},
		{name: "empty strategy", data: &price_calcular.PricingData{Strategy: nil}, want: false},
		{
			name: "single free plan",
			data: &price_calcular.PricingData{Strategy: map[string]*price_calcular.PricePlan{
				"r1": {Sale: map[string]float64{"per_call": 0}},
			}},
			want: true,
		},
		{
			name: "paid plan only",
			data: &price_calcular.PricingData{Strategy: map[string]*price_calcular.PricePlan{
				"r1": {Sale: map[string]float64{"per_call": 1.5}},
			}},
			want: false,
		},
		{
			name: "mixed plans",
			data: &price_calcular.PricingData{Strategy: map[string]*price_calcular.PricePlan{
				"r1": {Sale: map[string]float64{"per_call": 1.5}},
				"r2": {Sale: map[string]float64{"per_call": 0, "per_token": 0}},
			}},
			want: true,
		},
		{
			name: "nil plan and empty sale skipped",
			data: &price_calcular.PricingData{Strategy: map[string]*price_calcular.PricePlan{
				"r1": nil,
				"r2": {Sale: map[string]float64{}},
				"r3": {Sale: map[string]float64{"per_call": 2.0}},
			}},
			want: false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := hasFreePricePlan(c.data); got != c.want {
				t.Errorf("hasFreePricePlan(%s) = %v, want %v", c.name, got, c.want)
			}
		})
	}
}

// TestExecutor_ResetTypeCheck Reset 方法应校验入参类型
func TestExecutor_ResetTypeCheck(t *testing.T) {
	e := &executor{WorkerBase: drivers.Worker("id", "name")}
	if err := e.Reset("not a config", nil); err == nil {
		t.Errorf("expected type error for invalid config type, got nil")
	}
	cfg := &Config{Cache: "cache_id", ConcurrencyLimit: 200, EnableBalance: true}
	_ = checkConfig(cfg, nil)
	if err := e.Reset(cfg, nil); err != nil {
		t.Fatalf("Reset with valid Config returned error: %v", err)
	}
	if e.redisID != "cache_id" || e.defaultConcurrencyLimit != 200 || !e.enableBalance {
		t.Errorf("Reset did not apply config: %+v", e)
	}
	if e.balanceKeyGenerator == nil || e.priceKeyGenerator == nil ||
		e.taskKeyGenerator == nil || e.concurrencyKeyGenerator == nil {
		t.Errorf("Reset did not initialize key generators")
	}
}

// TestCreate 工厂函数应成功创建 executor 并完成基础初始化
func TestCreate(t *testing.T) {
	cfg := &Config{Cache: "cache_id", ConcurrencyLimit: 10, EnableBalance: false}
	_ = checkConfig(cfg, nil)
	w, err := Create("id_x", "name_x", cfg, nil)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	e, ok := w.(*executor)
	if !ok {
		t.Fatalf("expected *executor, got %T", w)
	}
	if e.redisID != "cache_id" || e.defaultConcurrencyLimit != 10 || e.enableBalance {
		t.Errorf("Create did not apply config: %+v", e)
	}
}

// TestExecutor_LifecycleAndSkill 覆盖 Start/Stop/Destroy/CheckSkill
func TestExecutor_LifecycleAndSkill(t *testing.T) {
	e := &executor{WorkerBase: drivers.Worker("id", "name")}
	if err := e.Start(); err != nil {
		t.Errorf("Start error: %v", err)
	}
	if err := e.Stop(); err != nil {
		t.Errorf("Stop error: %v", err)
	}
	e.Destroy() // 无返回值，仅确保不 panic

	if !e.CheckSkill(http_context.FilterSkillName) {
		t.Errorf("CheckSkill(%q) should be true", http_context.FilterSkillName)
	}
	if e.CheckSkill("unknown-skill") {
		t.Errorf("CheckSkill(unknown) should be false")
	}
}

// TestDoHttpFilter_MissingLabels 缺少 application/api 标签时应直接放行
func TestDoHttpFilter_MissingLabels(t *testing.T) {
	plugin := &executor{
		WorkerBase:              drivers.Worker("id", "name"),
		balanceKeyGenerator:     context_label.NewKeyGenerator("balance:{application}"),
		priceKeyGenerator:       context_label.NewKeyGenerator("price:{application}"),
		taskKeyGenerator:        context_label.NewKeyGenerator("task:{application}"),
		concurrencyKeyGenerator: context_label.NewKeyGenerator("concurrency:{application}"),
	}

	// application 缺失
	ctx := newMockHttpContext(map[string]string{"api": "api_1"}, &mockResponse{})
	if err := plugin.DoHttpFilter(ctx, &mockChain{}); err != nil {
		t.Errorf("expected nil error when application missing, got %v", err)
	}
	// api 缺失
	ctx = newMockHttpContext(map[string]string{"application": "app_1"}, &mockResponse{})
	if err := plugin.DoHttpFilter(ctx, &mockChain{}); err != nil {
		t.Errorf("expected nil error when api missing, got %v", err)
	}
}

// TestDoHttpFilter_NoResource 缺少 resource 标签时应跳过计费并走链路
func TestDoHttpFilter_NoResource(t *testing.T) {
	cache := &mockCache{kv: map[string]string{}}
	scope_manager.Set("mock_cache_noresource", cache, "redis")
	defer scope_manager.Del("mock_cache_noresource")

	plugin := &executor{
		WorkerBase:              drivers.Worker("id", "name"),
		redisID:                 "mock_cache_noresource",
		enableBalance:           false,
		balanceKeyGenerator:     context_label.NewKeyGenerator("balance:{application}"),
		priceKeyGenerator:       context_label.NewKeyGenerator("price:{application}"),
		taskKeyGenerator:        context_label.NewKeyGenerator("task:{application}"),
		concurrencyKeyGenerator: context_label.NewKeyGenerator("concurrency:{application}"),
	}
	ctx := newMockHttpContext(map[string]string{
		"application": "app_1",
		"api":         "api_1",
	}, &mockResponse{})
	if err := plugin.DoHttpFilter(ctx, &mockChain{}); err != nil {
		t.Errorf("expected nil error when resource missing, got %v", err)
	}
	// 未命中 resource 时不应产生任何扣款 Lua 调用
	if len(cache.runCalls) != 0 {
		t.Errorf("expected no run calls, got %d", len(cache.runCalls))
	}
}

// TestDoHttpFilter_CalculatorNotFound 计算器缺失时应放行链路且不扣费
func TestDoHttpFilter_CalculatorNotFound(t *testing.T) {
	cache := &mockCache{kv: map[string]string{}}
	scope_manager.Set("mock_cache_nocalc", cache, "redis")
	defer scope_manager.Del("mock_cache_nocalc")

	plugin := &executor{
		WorkerBase:              drivers.Worker("id", "name"),
		redisID:                 "mock_cache_nocalc",
		enableBalance:           false,
		balanceKeyGenerator:     context_label.NewKeyGenerator("balance:{application}"),
		priceKeyGenerator:       context_label.NewKeyGenerator("price:{application}"),
		taskKeyGenerator:        context_label.NewKeyGenerator("task:{application}"),
		concurrencyKeyGenerator: context_label.NewKeyGenerator("concurrency:{application}"),
	}
	ctx := newMockHttpContext(map[string]string{
		"application":   "app_1",
		"api":           "api_1",
		"resource":      "not_registered_res",
		"resource_type": "api",
	}, &mockResponse{})
	if err := plugin.DoHttpFilter(ctx, &mockChain{}); err != nil {
		t.Errorf("expected nil error when calculator missing, got %v", err)
	}
	if len(cache.runCalls) != 0 {
		t.Errorf("expected no run calls, got %d", len(cache.runCalls))
	}
}

// concurrencyCache 在 IncrBy 时始终返回超出阈值的计数，模拟并发触顶
type concurrencyCache struct {
	mockCache
	incrVal int64
}

func (c *concurrencyCache) IncrBy(ctx context.Context, key string, decrement int64, expiration time.Duration) resources.IntResult {
	return &mockIntResult{val: c.incrVal, err: nil}
}

// TestDoHttpFilter_ConcurrencyLimitExceeded 并发上限触顶应返回 429 且不进入计费
func TestDoHttpFilter_ConcurrencyLimitExceeded(t *testing.T) {
	price_calcular.SetCalculator("api:res_conc", newTestCalculator(t))
	defer price_calcular.DelCalculator("api:res_conc")

	cache := &concurrencyCache{
		mockCache: mockCache{kv: map[string]string{}},
		incrVal:   999, // 显著超过默认 1
	}
	scope_manager.Set("mock_cache_conc", cache, "redis")
	defer scope_manager.Del("mock_cache_conc")

	plugin := &executor{
		WorkerBase:              drivers.Worker("id", "name"),
		redisID:                 "mock_cache_conc",
		defaultConcurrencyLimit: 1,
		enableBalance:           false,
		balanceKeyGenerator:     context_label.NewKeyGenerator("balance:{application}"),
		priceKeyGenerator:       context_label.NewKeyGenerator("price:{application}"),
		taskKeyGenerator:        context_label.NewKeyGenerator("task:{application}"),
		concurrencyKeyGenerator: context_label.NewKeyGenerator("concurrency:{application}:{resource}"),
	}
	resp := &mockResponse{}
	ctx := newMockHttpContext(map[string]string{
		"application":   "app_1",
		"api":           "api_1",
		"resource":      "res_conc",
		"resource_type": "api",
	}, resp)
	if err := plugin.DoHttpFilter(ctx, &mockChain{}); err != nil {
		t.Errorf("expected nil error when concurrency limit exceeded, got %v", err)
	}
	if resp.statusCode != http.StatusTooManyRequests {
		t.Errorf("expected status 429, got %d", resp.statusCode)
	}
}
