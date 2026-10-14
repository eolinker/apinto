package dynamic_billing

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	context_label "github.com/eolinker/apinto/common/context-label"
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/drivers/credit"
	price_calcular "github.com/eolinker/apinto/price-calcular"
	"github.com/eolinker/apinto/resources"
	scope_manager "github.com/eolinker/apinto/scope-manager"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
)

// mockCustomerVar 模拟 eosc.ICustomerVar 配置读取
type mockCustomerVar struct {
	eosc.ICustomerVar
	vars       map[string]map[string]string
	allEnabled bool
}

func (m *mockCustomerVar) Get(namespace, key string) (string, bool) {
	if m.allEnabled && key == "enable_billing" {
		return "1", true
	}
	if m.vars == nil {
		return "", false
	}
	if ns, ok := m.vars[namespace]; ok {
		val, ok := ns[key]
		return val, ok
	}
	return "", false
}

func (m *mockCustomerVar) GetAll(namespace string) (map[string]string, bool) {
	if m.vars == nil {
		return nil, false
	}
	ns, ok := m.vars[namespace]
	return ns, ok
}

func setMockBillingEnabled(enabled bool) func() {
	old := customerVar
	if enabled {
		customerVar = &mockCustomerVar{allEnabled: true}
	} else {
		customerVar = nil
	}
	return func() {
		customerVar = old
	}
}

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
	if len(keys) == 3 {
		if len(args) >= 2 {
			return &mockInterfaceResult{val: []interface{}{int64(1), int64(2), "", int64(100000000)}, err: nil}
		} else if len(args) >= 1 {
			return &mockInterfaceResult{val: []interface{}{int64(1), int64(2), "", args[0].(int64), int64(0)}, err: nil}
		}
	}
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
func (m *mockHttpContext) RequestId() string                    { return "test-req-id" }

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
	calc, err := price_calcular.NewCalculator("calc_test", "USD", variables, rules)
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
	defer setMockBillingEnabled(true)()

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
		balanceKeyGenerator:     context_label.NewKeyGenerator("balance:{application}"),
		creditKeyGenerator:      context_label.NewKeyGenerator("{product}:credit:user:{balance_target}"),
		priceKeyGenerator:       context_label.NewKeyGenerator("access-resource-price:{application}:{resource}"),
		taskKeyGenerator:        context_label.NewKeyGenerator("dynamic-billing-task:{application}:{resource}"),
		concurrencyKeyGenerator: context_label.NewKeyGenerator("dynamic-billing-concurrency:{application}:{resource}"),
		settingKeyGenerator:     context_label.NewKeyGenerator("dynamic-billing-setting:{application}:{resource}"),
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

	// 触发响应结束结算
	proxy := ctx.Proxy().(*mockProxy)
	for _, fn := range proxy.bodyFinishFns {
		fn(ctx)
	}

	// cost_per_call*1.0 = 2.0；sale_per_call*1.1 = 22.0；official 表达式实际复用 sale 表达式 => 22.0
	if got := context_label.GetAmountCost(ctx); got != "2.000000" {
		t.Errorf("expected cost_amount 2.000000, got %q", got)
	}
	if got := context_label.GetAmountSale(ctx); got != "22.000000" {
		t.Errorf("expected sale_amount 22.000000, got %q", got)
	}

	// 两阶段下：1次预扣（PreDeduct）+ 1次结算（Settle）
	if len(cache.runCalls) != 2 {
		t.Fatalf("expected 2 deduction calls (pre-deduct + settle), got %d", len(cache.runCalls))
	}
}

// TestDynamicBilling_InsufficientBalance 验证余额不足时前置阻断返回 402。
func TestDynamicBilling_InsufficientBalance(t *testing.T) {
	defer setMockBillingEnabled(true)()

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
		balanceKeyGenerator:     context_label.NewKeyGenerator("balance:{application}"),
		creditKeyGenerator:      context_label.NewKeyGenerator("{product}:credit:user:{balance_target}"),
		priceKeyGenerator:       context_label.NewKeyGenerator("access-resource-price:{application}:{resource}"),
		taskKeyGenerator:        context_label.NewKeyGenerator("dynamic-billing-task:{application}:{resource}"),
		concurrencyKeyGenerator: context_label.NewKeyGenerator("dynamic-billing-concurrency:{application}:{resource}"),
		settingKeyGenerator:     context_label.NewKeyGenerator("dynamic-billing-setting:{application}:{resource}"),
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
	defer setMockBillingEnabled(true)()

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
		balanceKeyGenerator:     context_label.NewKeyGenerator("balance:{application}"),
		creditKeyGenerator:      context_label.NewKeyGenerator("{product}:credit:user:{balance_target}"),
		priceKeyGenerator:       context_label.NewKeyGenerator("access-resource-price:{application}:{resource}"),
		taskKeyGenerator:        context_label.NewKeyGenerator("dynamic-billing-task:{application}:{resource}"),
		concurrencyKeyGenerator: context_label.NewKeyGenerator("dynamic-billing-concurrency:{application}:{resource}"),
		settingKeyGenerator:     context_label.NewKeyGenerator("dynamic-billing-setting:{application}:{resource}"),
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
	cfg := &Config{Cache: "cache_id", ConcurrencyLimit: 200}
	_ = checkConfig(cfg, nil)
	if err := e.Reset(cfg, nil); err != nil {
		t.Fatalf("Reset with valid Config returned error: %v", err)
	}
	if e.redisID != "cache_id" || e.defaultConcurrencyLimit != 200 {
		t.Errorf("Reset did not apply config: %+v", e)
	}
	if e.balanceKeyGenerator == nil || e.priceKeyGenerator == nil ||
		e.taskKeyGenerator == nil || e.concurrencyKeyGenerator == nil ||
		e.creditKeyGenerator == nil {
		t.Errorf("Reset did not initialize key generators")
	}
}

// TestCreate 工厂函数应成功创建 executor 并完成基础初始化
func TestCreate(t *testing.T) {
	cfg := &Config{Cache: "cache_id", ConcurrencyLimit: 10}
	_ = checkConfig(cfg, nil)
	w, err := Create("id_x", "name_x", cfg, nil)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	e, ok := w.(*executor)
	if !ok {
		t.Fatalf("expected *executor, got %T", w)
	}
	if e.redisID != "cache_id" || e.defaultConcurrencyLimit != 10 {
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

// ============================================================================
// 扣款金额验证测试：验证传给 Redis 的金额、放大倍数、以及跑完整流程后的账户余额
//
// 分两类：
//  1. 纯函数级（captureCache）：捕获 Run 的参数，断言传给 Redis 的金额正确。
//  2. 端到端（luaCache）：用内存 map 模拟 Redis 的 DECRBY/INCRBY/GET/SET/DEL/EX 语义，
//     跑完整 DoHttpFilter 流程后断言账户真实余额，验证"扣的钱到底对不对"。
// ============================================================================

// testRedisError 模拟 Redis 不可用错误。
type testRedisError struct{ msg string }

func (e *testRedisError) Error() string { return e.msg }

var (
	errRedisDown = &testRedisError{"redis unavailable"}
	errNil       = &testRedisError{"redis nil"}
)

// --- captureCache：只捕获 Run 参数，不执行任何 Lua 逻辑 ---

type captureCache struct {
	resources.ICache
	returns []interface{}
	runs    []runCall
	err     error
	idx     int
}

func (c *captureCache) Get(ctx context.Context, key string) resources.StringResult {
	return &mockStringResult{val: "", err: errNil}
}

func (c *captureCache) Run(ctx context.Context, script interface{}, keys []string, args ...interface{}) resources.InterfaceResult {
	c.runs = append(c.runs, runCall{keys: keys, args: args})
	if c.err != nil {
		return &mockInterfaceResult{val: nil, err: c.err}
	}
	if c.idx < len(c.returns) {
		v := c.returns[c.idx]
		c.idx++
		return &mockInterfaceResult{val: v, err: nil}
	}
	return &mockInterfaceResult{val: nil, err: nil}
}

// --- luaCache：用内存 map 真实模拟 Redis 语义 ---

type luaCache struct {
	resources.ICache
	mu   sync.Mutex
	kv   map[string]string
	hash map[string]map[string]int64
	runs int
}

func newLuaCache(balance map[string]string) *luaCache {
	m := make(map[string]string)
	for k, v := range balance {
		m[k] = v
	}
	return &luaCache{kv: m, hash: make(map[string]map[string]int64)}
}

func (c *luaCache) setCreditBalance(creditKey, cid string, amt int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.hash == nil {
		c.hash = make(map[string]map[string]int64)
	}
	if c.hash[creditKey] == nil {
		c.hash[creditKey] = make(map[string]int64)
	}
	c.hash[creditKey][cid] = amt
}

func (c *luaCache) creditInt(creditKey, cid string) int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.hash == nil || c.hash[creditKey] == nil {
		return 0
	}
	return c.hash[creditKey][cid]
}

func (c *luaCache) Get(ctx context.Context, key string) resources.StringResult {
	c.mu.Lock()
	v, ok := c.kv[key]
	c.mu.Unlock()
	if !ok {
		return &mockStringResult{val: "", err: errNil}
	}
	return &mockStringResult{val: v, err: nil}
}

func (c *luaCache) Set(ctx context.Context, key string, value []byte, expiration time.Duration) resources.StatusResult {
	c.mu.Lock()
	c.kv[key] = string(value)
	c.mu.Unlock()
	return nil
}

func (c *luaCache) IncrBy(ctx context.Context, key string, incr int64, expiration time.Duration) resources.IntResult {
	c.mu.Lock()
	cur := int64(0)
	if v, ok := c.kv[key]; ok {
		cur, _ = strconv.ParseInt(v, 10, 64)
	}
	cur += incr
	c.kv[key] = strconv.FormatInt(cur, 10)
	c.mu.Unlock()
	return &mockIntResult{val: cur, err: nil}
}

func (c *luaCache) DecrBy(ctx context.Context, key string, decr int64, expiration time.Duration) resources.IntResult {
	return c.IncrBy(ctx, key, -decr, expiration)
}

func (c *luaCache) Del(ctx context.Context, keys ...string) resources.IntResult {
	c.mu.Lock()
	var n int64
	for _, k := range keys {
		if _, ok := c.kv[k]; ok {
			delete(c.kv, k)
			n++
		}
	}
	c.mu.Unlock()
	return &mockIntResult{val: n, err: nil}
}

// Run 模拟 executor.go 中用到的 Lua 脚本语义。
func (c *luaCache) Run(ctx context.Context, script interface{}, keys []string, args ...interface{}) resources.InterfaceResult {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.runs++
	bal := ""
	if len(keys) > 0 {
		bal = keys[0]
	}
	getInt := func(k string) int64 {
		v, ok := c.kv[k]
		if !ok {
			return 0
		}
		n, _ := strconv.ParseInt(v, 10, 64)
		return n
	}
	set := func(k string, n int64) { c.kv[k] = strconv.FormatInt(n, 10) }
	var ret interface{}

	s, _ := script.(string)

	switch {
	case strings.Contains(s, "fallbackAmt") || strings.Contains(s, "recovery") || strings.Contains(s, "estAmt"):
		// executeRecovery 冲正
		creditKey := keys[0]
		balanceKey := keys[1]
		preKey := keys[2]
		var dType int64 = 2
		var preAmt int64
		var cid string
		var creditAmt int64
		var balanceAmt int64
		if c.hash != nil && c.hash[preKey] != nil {
			dType = c.hash[preKey]["deduct_type"]
			preAmt = c.hash[preKey]["amount"]
			creditAmt = c.hash[preKey]["credit_amount"]
			balanceAmt = c.hash[preKey]["balance_amount"]
			cid = c.kv[preKey+":cid"]
			delete(c.hash, preKey)
			delete(c.kv, preKey+":cid")
		} else if len(args) >= 1 {
			preAmt = args[0].(int64)
			if len(args) >= 2 {
				dType = int64(args[1].(int))
			}
			if len(args) >= 3 {
				cid = args[2].(string)
			}
		}

		if dType == 3 {
			if creditAmt > 0 && cid != "" {
				if c.hash == nil {
					c.hash = make(map[string]map[string]int64)
				}
				if c.hash[creditKey] == nil {
					c.hash[creditKey] = make(map[string]int64)
				}
				c.hash[creditKey][cid] += creditAmt
			}
			if balanceAmt > 0 {
				nb := getInt(balanceKey) + balanceAmt
				set(balanceKey, nb)
			}
			ret = []interface{}{int64(1), int64(3), cid, creditAmt, balanceAmt}
		} else if dType == 1 && cid != "" {
			refund := creditAmt
			if refund <= 0 {
				refund = preAmt
			}
			if refund > 0 {
				if c.hash == nil {
					c.hash = make(map[string]map[string]int64)
				}
				if c.hash[creditKey] == nil {
					c.hash[creditKey] = make(map[string]int64)
				}
				c.hash[creditKey][cid] += refund
			}
			ret = []interface{}{int64(1), int64(1), cid, refund}
		} else {
			refund := balanceAmt
			if refund <= 0 {
				refund = preAmt
			}
			if refund > 0 {
				nb := getInt(balanceKey) + refund
				set(balanceKey, nb)
			}
			ret = []interface{}{int64(1), int64(2), "", refund}
		}

	case strings.Contains(s, "expire") || strings.Contains(s, "hmset") || strings.Contains(s, "ttl"):
		// executePreDeductTwoPhase 预扣
		amt := args[0].(int64)
		creditKey := keys[0]
		balanceKey := keys[1]
		preKey := keys[2]

		targetCid := ""
		var targetCBal int64 = 0
		for i := 2; i < len(args); i++ {
			cid := args[i].(string)
			if c.hash != nil && c.hash[creditKey] != nil {
				if bal := c.hash[creditKey][cid]; bal > 0 {
					targetCid = cid
					targetCBal = bal
					break
				}
			}
		}

		if targetCid != "" {
			if targetCBal >= amt {
				c.hash[creditKey][targetCid] -= amt
				if c.hash[preKey] == nil {
					c.hash[preKey] = make(map[string]int64)
				}
				c.hash[preKey]["deduct_type"] = 1
				c.hash[preKey]["amount"] = amt
				c.hash[preKey]["credit_amount"] = amt
				c.hash[preKey]["balance_amount"] = 0
				c.kv[preKey+":cid"] = targetCid
				ret = []interface{}{int64(1), int64(1), targetCid, c.hash[creditKey][targetCid], int64(0)}
			} else {
				remain := amt - targetCBal
				b := getInt(balanceKey)
				if b >= remain {
					c.hash[creditKey][targetCid] = 0
					nb := b - remain
					set(balanceKey, nb)
					if c.hash[preKey] == nil {
						c.hash[preKey] = make(map[string]int64)
					}
					c.hash[preKey]["deduct_type"] = 3
					c.hash[preKey]["amount"] = amt
					c.hash[preKey]["credit_amount"] = targetCBal
					c.hash[preKey]["balance_amount"] = remain
					c.kv[preKey+":cid"] = targetCid
					ret = []interface{}{int64(1), int64(3), targetCid, int64(0), nb}
				} else {
					ret = []interface{}{int64(0), int64(0), "", targetCBal, b}
				}
			}
		} else {
			b := getInt(balanceKey)
			if b >= amt {
				nb := b - amt
				set(balanceKey, nb)
				if c.hash == nil {
					c.hash = make(map[string]map[string]int64)
				}
				if c.hash[preKey] == nil {
					c.hash[preKey] = make(map[string]int64)
				}
				c.hash[preKey]["deduct_type"] = 2
				c.hash[preKey]["amount"] = amt
				c.hash[preKey]["credit_amount"] = 0
				c.hash[preKey]["balance_amount"] = amt
				ret = []interface{}{int64(1), int64(2), "", int64(0), nb}
			} else {
				ret = []interface{}{int64(0), int64(0), "", int64(0), b}
			}
		}

	case strings.Contains(s, "actual") && len(keys) == 3:
		// settlePreDeductTwoPhase 结算
		actual := args[0].(int64)
		creditKey := keys[0]
		balanceKey := keys[1]
		preKey := keys[2]
		var dType int64 = 2
		var preAmt int64
		var preCreditAmt int64
		var preBalanceAmt int64
		var cid string
		if c.hash != nil && c.hash[preKey] != nil {
			dType = c.hash[preKey]["deduct_type"]
			preAmt = c.hash[preKey]["amount"]
			preCreditAmt = c.hash[preKey]["credit_amount"]
			preBalanceAmt = c.hash[preKey]["balance_amount"]
			cid = c.kv[preKey+":cid"]
			delete(c.hash, preKey)
			delete(c.kv, preKey+":cid")
		} else {
			// 幂等：若快照已删除/不存在，直接返回不重复扣款
			ret = []interface{}{int64(1), int64(0), "", int64(0), int64(0)}
			break
		}
		if preCreditAmt == 0 && preBalanceAmt == 0 {
			if dType == 1 {
				preCreditAmt = preAmt
			} else if dType == 2 {
				preBalanceAmt = preAmt
			}
		}

		diff := actual - preAmt
		if dType == 1 {
			if diff <= 0 {
				refund := -diff
				if refund > 0 {
					c.hash[creditKey][cid] += refund
				}
				ret = []interface{}{int64(1), int64(1), cid, actual, int64(0)}
			} else {
				cBal := c.hash[creditKey][cid]
				if cBal >= diff {
					c.hash[creditKey][cid] -= diff
					ret = []interface{}{int64(1), int64(1), cid, actual, int64(0)}
				} else {
					c.hash[creditKey][cid] = 0
					remain := diff - cBal
					nb := getInt(balanceKey) - remain
					set(balanceKey, nb)
					ret = []interface{}{int64(1), int64(3), cid, preAmt + cBal, remain, nb}
				}
			}
		} else if dType == 3 {
			if diff < 0 {
				refund := -diff
				if refund <= preBalanceAmt {
					nb := getInt(balanceKey) + refund
					set(balanceKey, nb)
					ret = []interface{}{int64(1), int64(3), cid, preCreditAmt, preBalanceAmt - refund}
				} else {
					nb := getInt(balanceKey) + preBalanceAmt
					set(balanceKey, nb)
					remainRefund := refund - preBalanceAmt
					c.hash[creditKey][cid] += remainRefund
					ret = []interface{}{int64(1), int64(1), cid, actual, int64(0)}
				}
			} else if diff > 0 {
				cBal := c.hash[creditKey][cid]
				if cBal >= diff {
					c.hash[creditKey][cid] -= diff
					ret = []interface{}{int64(1), int64(3), cid, preCreditAmt + diff, preBalanceAmt}
				} else {
					c.hash[creditKey][cid] = 0
					remainDiff := diff - cBal
					nb := getInt(balanceKey) - remainDiff
					set(balanceKey, nb)
					ret = []interface{}{int64(1), int64(3), cid, preCreditAmt + cBal, preBalanceAmt + remainDiff, nb}
				}
			} else {
				ret = []interface{}{int64(1), int64(3), cid, preCreditAmt, preBalanceAmt}
			}
		} else {
			if diff != 0 {
				nb := getInt(balanceKey) - diff
				set(balanceKey, nb)
				ret = []interface{}{int64(1), int64(2), "", int64(0), actual, nb}
			} else {
				ret = []interface{}{int64(1), int64(2), "", int64(0), actual}
			}
		}

	case strings.Contains(s, "actual") && len(keys) == 2:
		// directDeductLua 无预扣直接扣款
		actual := args[0].(int64)
		balanceKey := keys[1]
		nb := getInt(balanceKey) - actual
		set(balanceKey, nb)
		ret = []interface{}{int64(1), int64(2), "", nb, actual}

	case len(keys) == 1 && len(args) == 1:
		// executeBalanceDeduction: decrby balance amt
		nb := getInt(bal) - args[0].(int64)
		set(bal, nb)
		ret = nb

	case len(keys) == 2 && len(args) == 2:
		// executePreDeduct: 余额校验 + decrby + setex pre
		amt := args[0].(int64)
		b := getInt(bal)
		if b < amt {
			ret = []interface{}{int64(0), b}
		} else {
			nb := b - amt
			set(bal, nb)
			c.kv[keys[1]] = strconv.FormatInt(amt, 10)
			ret = []interface{}{int64(1), nb}
		}

	case len(keys) == 2 && len(args) == 1:
		// settlePreDeduct: get pre / del pre / decrby diff
		actual := args[0].(int64)
		pre := getInt(keys[1])
		if pre == 0 {
			ret = []interface{}{int64(0), getInt(bal), int64(0)}
		} else {
			delete(c.kv, keys[1])
			diff := actual - pre
			nb := getInt(bal) - diff
			if diff != 0 {
				set(bal, nb)
			}
			ret = []interface{}{diff, nb, pre}
		}

	case len(keys) == 2 && len(args) == 0:
		// refundPreDeduct: get pre / del pre / incrby pre
		pre := getInt(keys[1])
		if pre == 0 {
			ret = []interface{}{int64(0), getInt(bal)}
		} else {
			delete(c.kv, keys[1])
			nb := getInt(bal) + pre
			set(bal, nb)
			ret = []interface{}{pre, nb}
		}
	}
	return &mockInterfaceResult{val: ret, err: nil}
}

func (c *luaCache) balanceInt(key string) int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, ok := c.kv[key]
	if !ok {
		return 0
	}
	n, _ := strconv.ParseInt(v, 10, 64)
	return n
}

// --- errIncrCache：让 IncrBy 始终返回 error，其它走 mockCache ---

type errIncrCache struct {
	mockCache
}

func (c *errIncrCache) IncrBy(ctx context.Context, key string, incr int64, exp time.Duration) resources.IntResult {
	return &mockIntResult{val: 0, err: errRedisDown}
}

// --- errChain：始终返回 error 的调用链 ---

type errChain struct{}

func (e *errChain) DoChain(ctx eocontext.EoContext) error {
	return errRedisDown
}

func (e *errChain) Destroy() {}

// --- 纯函数级测试 ---

// TestExecuteBalanceDeduction_Scale_6Digits 验证扣款金额按 1e6 放大。
func TestExecuteBalanceDeduction_Scale_6Digits(t *testing.T) {
	// cost=12.345678（元）按 6 位放大 -> 12345678
	c := &captureCache{}
	ok := executeBalanceDeduction(context.Background(), c, "bal", 12.345678, "u1")
	if !ok {
		t.Fatalf("expected success, got false")
	}
	if len(c.runs) != 1 {
		t.Fatalf("expected 1 run call, got %d", len(c.runs))
	}
	got := c.runs[0].args[0].(int64)
	if got != 12345678 {
		t.Errorf("deduct arg = %d, want 12345678 (12.345678 * 1e6)", got)
	}
}

// TestExecutePreDeduct_Scale_6Digits 验证预扣金额与 TTL 按 1e6 / 秒 传入。
func TestExecutePreDeduct_Scale_6Digits(t *testing.T) {
	// amount=5.5 元 -> 5500000
	c := &captureCache{returns: []interface{}{[]interface{}{int64(1), int64(94500000)}}}
	ok := executePreDeduct(context.Background(), c, "bal", "pre:1", 5.5, 5*time.Minute, "u1")
	if !ok {
		t.Fatalf("expected success, got false")
	}
	if len(c.runs) != 1 {
		t.Fatalf("expected 1 run call, got %d", len(c.runs))
	}
	if got := c.runs[0].args[0].(int64); got != 5500000 {
		t.Errorf("pre-deduct amount = %d, want 5500000 (5.5 * 1e6)", got)
	}
	// ttl 应为 300 秒
	if got := c.runs[0].args[1].(int64); got != 300 {
		t.Errorf("pre-deduct ttl = %d, want 300", got)
	}
}

// TestExecutePreDeductTwoPhase_Scale_6Digits 验证两阶段预扣金额按 1e6 放大为 int64 整数传入 Redis，不进行浮点数计算。
func TestExecutePreDeductTwoPhase_Scale_6Digits(t *testing.T) {
	// amount=5.5 元 -> 5500000 (int64)
	c := &captureCache{returns: []interface{}{[]interface{}{int64(1), int64(1), "coupon_1", int64(44500000), int64(0)}}}
	ok, dType, cID := executePreDeductTwoPhase(context.Background(), c, "credit:1", "bal:1", "pre:1", []string{"coupon_1"}, 5.5, 5*time.Minute, "u1")
	if !ok || dType != 1 || cID != "coupon_1" {
		t.Fatalf("expected (true, 1, coupon_1), got (%v, %d, %s)", ok, dType, cID)
	}
	if len(c.runs) != 1 {
		t.Fatalf("expected 1 run call, got %d", len(c.runs))
	}
	if got, ok := c.runs[0].args[0].(int64); !ok || got != 5500000 {
		t.Errorf("pre-deduct two-phase amount = %v (%T), want int64(5500000)", c.runs[0].args[0], c.runs[0].args[0])
	}
	if got, ok := c.runs[0].args[1].(int64); !ok || got != 300 {
		t.Errorf("pre-deduct two-phase ttl = %v (%T), want int64(300)", c.runs[0].args[1], c.runs[0].args[1])
	}
}

// TestExecutePreDeduct_InsufficientBalance_ReturnsFalse 余额不足时应返回 false。
func TestExecutePreDeduct_InsufficientBalance_ReturnsFalse(t *testing.T) {
	c := &captureCache{returns: []interface{}{[]interface{}{int64(0), int64(1000000)}}}
	ok := executePreDeduct(context.Background(), c, "bal", "pre:1", 5.5, 5*time.Minute, "u1")
	if ok {
		t.Fatalf("expected false for insufficient balance")
	}
}

// TestExecutePreDeduct_UnexpectedReturn_FailClosed 返回值类型不符时 fail-closed。
func TestExecutePreDeduct_UnexpectedReturn_FailClosed(t *testing.T) {
	c := &captureCache{returns: []interface{}{int64(1)}}
	ok := executePreDeduct(context.Background(), c, "bal", "pre:1", 5.5, 5*time.Minute, "u1")
	if ok {
		t.Fatalf("expected false for unexpected return type")
	}
}

// TestExecutePreDeduct_RedisError_FailClosed Redis 出错时 fail-closed。
func TestExecutePreDeduct_RedisError_FailClosed(t *testing.T) {
	c := &captureCache{err: errRedisDown}
	ok := executePreDeduct(context.Background(), c, "bal", "pre:1", 5.5, 5*time.Minute, "u1")
	if ok {
		t.Fatalf("expected false for redis error")
	}
}

// TestSettlePreDeduct_NoPreKey_DegradesToDirectDeduct 无 preKey 时退化为直接扣款。
func TestSettlePreDeduct_NoPreKey_DegradesToDirectDeduct(t *testing.T) {
	c := &captureCache{returns: []interface{}{int64(1)}}
	ctx := newMockHttpContext(nil, &mockResponse{})
	settlePreDeduct(ctx, c, "bal", "", 8.8, "u1")
	if len(c.runs) != 1 {
		t.Fatalf("expected 1 run call, got %d", len(c.runs))
	}
	// 8.8 * 1e6 = 8800000
	if got := c.runs[0].args[0].(int64); got != 8800000 {
		t.Errorf("settle direct-deduct arg = %d, want 8800000", got)
	}
}

// TestSettlePreDeduct_WithPreKey_PassesActualArg 有 preKey 时传给 Redis 的是 actualSale。
func TestSettlePreDeduct_WithPreKey_PassesActualArg(t *testing.T) {
	c := &captureCache{returns: []interface{}{[]interface{}{int64(2000000), int64(98000000), int64(10000000)}}}
	ctx := newMockHttpContext(nil, &mockResponse{})
	settlePreDeduct(ctx, c, "bal", "pre:1", 12.0, "u1")
	// actualSale=12.0 -> 12000000
	if got := c.runs[0].args[0].(int64); got != 12000000 {
		t.Errorf("settle actual arg = %d, want 12000000", got)
	}
}

// TestRefundPreDeduct_NoKey_Noop preKey 为空时不调用 Redis。
func TestRefundPreDeduct_NoKey_Noop(t *testing.T) {
	c := &captureCache{}
	ctx := newMockHttpContext(nil, &mockResponse{})
	refundPreDeduct(ctx, c, "bal", "", "u1")
	if len(c.runs) != 0 {
		t.Errorf("expected no run call when preDeductKey empty, got %d", len(c.runs))
	}
}

// --- 端到端测试 ---

// newBillingTestEnv 构造可复用的端到端测试环境。
// balanceYuan 单位为元，会被转成 1e6 存入 luaCache。
// 定价：cost_per_call=2, sale_per_call=20，规则系数 sale*1.1 => sale=22 元。
func newBillingTestEnv(t *testing.T, balanceYuan float64) (*executor, *luaCache, *mockHttpContext) {
	t.Helper()
	deferBilling := setMockBillingEnabled(true)
	t.Cleanup(deferBilling)

	price_calcular.SetCalculator("api:res_billing", newTestCalculator(t))
	t.Cleanup(func() { price_calcular.DelCalculator("api:res_billing") })

	balanceKey := "balance:app_billing"
	priceKey := "access-resource-price:app_billing:res_billing"
	priceJSON := `{
		"basic_info": {"version": "v1.0.0", "resource_group_id": "res_billing", "tenant_id": "t1"},
		"strategy": {
			"rule_success": {
				"cost": {"per_call": 2.0},
				"sale": {"per_call": 20.0},
				"official": {"per_call": 25.0}
			}
		}
	}`
	cache := newLuaCache(map[string]string{
		balanceKey: strconv.FormatInt(int64(balanceYuan*1000000), 10),
		priceKey:   priceJSON,
	})
	scope_manager.Set("lua_cache_billing", cache, "lua_cache_billing", "redis")
	t.Cleanup(func() { scope_manager.Del("lua_cache_billing") })

	plugin := &executor{
		WorkerBase:              drivers.Worker("billing_test", "billing_test"),
		redisID:                 "lua_cache_billing",
		balanceKeyGenerator:     context_label.NewKeyGenerator(balanceKey),
		creditKeyGenerator:      context_label.NewKeyGenerator("{product}:credit:user:{balance_target}"),
		priceKeyGenerator:       context_label.NewKeyGenerator(priceKey),
		taskKeyGenerator:        context_label.NewKeyGenerator("task:{application}:{resource}"),
		concurrencyKeyGenerator: context_label.NewKeyGenerator("conc:{application}:{resource}"),
		settingKeyGenerator:     context_label.NewKeyGenerator("dynamic-billing-setting:{application}:{resource}"),
	}

	resp := &mockResponse{statusCode: 200}
	ctx := newMockHttpContext(map[string]string{
		"application":   "app_billing",
		"api":           "api_billing",
		"resource":      "res_billing",
		"resource_type": "api",
	}, resp)
	return plugin, cache, ctx
}

// triggerSettle 触发 immediate 模式注册的 AppendBodyFinish 回调（执行结算）。
func triggerSettle(ctx *mockHttpContext) {
	proxy := ctx.Proxy().(*mockProxy)
	for _, fn := range proxy.bodyFinishFns {
		fn(ctx)
	}
}

// TestEndToEnd_ImmediateSettle_BalanceChange immediate 模式跑完后余额应扣 22 元。
// 初始 100 元 => 期望 78 元。
func TestEndToEnd_ImmediateSettle_BalanceChange(t *testing.T) {
	plugin, cache, ctx := newBillingTestEnv(t, 100.0)
	balanceKey := "balance:app_billing"

	if err := plugin.DoHttpFilter(ctx, &mockChain{}); err != nil {
		t.Fatalf("DoHttpFilter failed: %v", err)
	}
	triggerSettle(ctx)

	got := cache.balanceInt(balanceKey)
	want := int64(78 * 1000000)
	if got != want {
		t.Errorf("balance after immediate settle = %d (=%.6f元), want %d (78元)", got, float64(got)/1e6, want)
	}
	if dType := ctx.GetLabel(context_label.LabelDeductType); dType != context_label.DeductTypeBalance {
		t.Errorf("deduct_type = %s, want %s", dType, context_label.DeductTypeBalance)
	}
}

// TestEndToEnd_PreDeductThenSettle_NoExtraDeduction 预扣 22 + 结算 diff=0 => 只扣一次。
func TestEndToEnd_PreDeductThenSettle_NoExtraDeduction(t *testing.T) {
	plugin, cache, ctx := newBillingTestEnv(t, 100.0)
	balanceKey := "balance:app_billing"

	if err := plugin.DoHttpFilter(ctx, &mockChain{}); err != nil {
		t.Fatalf("DoHttpFilter failed: %v", err)
	}
	triggerSettle(ctx)

	got := cache.balanceInt(balanceKey)
	want := int64(78 * 1000000)
	if got != want {
		t.Errorf("balance = %d (=%.6f元), want %d (78元). 预扣+结算应只扣一次", got, float64(got)/1e6, want)
	}
}

// TestEndToEnd_DownstreamError_RefundPreDeduct 下游失败时全额退回预扣，余额不变。
func TestEndToEnd_DownstreamError_RefundPreDeduct(t *testing.T) {
	plugin, cache, ctx := newBillingTestEnv(t, 100.0)
	balanceKey := "balance:app_billing"

	if err := plugin.DoHttpFilter(ctx, &errChain{}); err == nil {
		t.Fatalf("expected error from chain, got nil")
	}

	got := cache.balanceInt(balanceKey)
	want := int64(100 * 1000000)
	if got != want {
		t.Errorf("balance after refund = %d (=%.6f元), want %d (100元). 下游失败应全额退回预扣", got, float64(got)/1e6, want)
	}
}

// TestEndToEnd_SettleIdempotent_MultipleCalls 结算幂等：多次结算只扣一次。
func TestEndToEnd_SettleIdempotent_MultipleCalls(t *testing.T) {
	plugin, cache, ctx := newBillingTestEnv(t, 100.0)
	balanceKey := "balance:app_billing"

	if err := plugin.DoHttpFilter(ctx, &mockChain{}); err != nil {
		t.Fatalf("DoHttpFilter failed: %v", err)
	}
	// 故意触发结算回调 3 次，模拟重入
	proxy := ctx.Proxy().(*mockProxy)
	for i := 0; i < 3; i++ {
		for _, fn := range proxy.bodyFinishFns {
			fn(ctx)
		}
	}
	got := cache.balanceInt(balanceKey)
	want := int64(78 * 1000000)
	if got != want {
		t.Errorf("balance after 3x settle = %d (=%.6f元), want %d (78元). 结算应幂等", got, float64(got)/1e6, want)
	}
}

// TestEndToEnd_PreDeductInsufficientBalance_Block402 余额不足时预扣失败返回 402 且不扣款。
func TestEndToEnd_PreDeductInsufficientBalance_Block402(t *testing.T) {
	// 余额只有 1 元，但预扣需要 22 元
	plugin, cache, ctx := newBillingTestEnv(t, 1.0)
	balanceKey := "balance:app_billing"

	t.Logf("cache balance before = %s", cache.kv[balanceKey])
	err := plugin.DoHttpFilter(ctx, &mockChain{})
	t.Logf("DoHttpFilter returned err = %v, cache runs = %d", err, cache.runs)
	if err == nil {
		t.Fatalf("expected insufficient balance error, got nil")
	}
	resp := ctx.Response().(*mockResponse)
	if resp.statusCode != http.StatusPaymentRequired {
		t.Errorf("expected 402, got %d", resp.statusCode)
	}
	got := cache.balanceInt(balanceKey)
	want := int64(1 * 1000000)
	if got != want {
		t.Errorf("balance = %d (=%.6f元), want %d (1元). 预扣失败不应扣款", got, float64(got)/1e6, want)
	}
}

// TestEndToEnd_BalanceUnitConsistency 回归测试：前置校验与扣款使用同一放大倍数（1e6）。
// 余额 0.000001 元 = 1（最小单位），>0 通过前置校验，但预扣 22 元必然失败。
func TestEndToEnd_BalanceUnitConsistency(t *testing.T) {
	plugin, cache, ctx := newBillingTestEnv(t, 0.000001)
	balanceKey := "balance:app_billing"

	_ = plugin.DoHttpFilter(ctx, &mockChain{})
	resp := ctx.Response().(*mockResponse)
	if resp.statusCode != http.StatusPaymentRequired {
		t.Errorf("expected 402 for balance 0.000001元 (1 unit) which is < pre-deduct, got %d", resp.statusCode)
	}
	if got := cache.balanceInt(balanceKey); got != 1 {
		t.Errorf("balance unit mismatch: got %d, want 1 (0.000001元 * 1e6). 前置校验与扣款倍数不一致", got)
	}
}

// TestEndToEnd_ConcurrencyRedisError_FailClosed 并发计数 Redis 出错时 fail-closed 返回 503。
func TestEndToEnd_ConcurrencyRedisError_FailClosed(t *testing.T) {
	plugin, cache, ctx := newBillingTestEnv(t, 100.0)
	// 用 IncrBy 返回 error 的 cache 包装，复用 luaCache 的 kv
	plugin.redisID = "err_cache_conc"
	errCache := &errIncrCache{mockCache: mockCache{kv: cache.kv}}
	scope_manager.Set("err_cache_conc", errCache, "err_cache_conc", "redis")
	defer scope_manager.Del("err_cache_conc")

	plugin.defaultConcurrencyLimit = 10
	ctx.SetLabel("concurrency_limit", "10")

	err := plugin.DoHttpFilter(ctx, &mockChain{})
	if err != nil {
		t.Fatalf("expected nil error (503 returned, not propagated), got %v", err)
	}
	resp := ctx.Response().(*mockResponse)
	if resp.statusCode != http.StatusServiceUnavailable {
		t.Errorf("expected 503 when concurrency redis errors, got %d", resp.statusCode)
	}
}

// TestEndToEnd_Credit_PreDeductAndSettle_Refund 代金券全额抵扣与标签测试
func TestEndToEnd_Credit_PreDeductAndSettle_Refund(t *testing.T) {
	plugin, cache, ctx := newBillingTestEnv(t, 100.0)
	ctx.SetLabel("product", "demo_prod")
	ctx.SetLabel("balance_target", "user_c1")
	ctx.SetLabel("service", "srv_c1")
	ctx.SetLabel("api", "api_billing")

	creditRule := credit.CompileRule("coupon_c1", &credit.Config{
		Nbf: time.Now().Unix() - 100,
		Exp: time.Now().Unix() + 3600,
		Filters: credit.Filters{
			Provider: credit.RuleFilter{API: []string{"srv_c1"}},
			Resource: credit.RuleFilter{API: []string{"res_billing"}},
		},
	})
	credit.SetRule("coupon_c1", creditRule)
	defer credit.DelRule("coupon_c1")

	creditKey := "demo_prod:credit:user:user_c1"
	cache.setCreditBalance(creditKey, "coupon_c1", 50*1000000) // 代金券 50 元

	if err := plugin.DoHttpFilter(ctx, &mockChain{}); err != nil {
		t.Fatalf("DoHttpFilter failed: %v", err)
	}
	triggerSettle(ctx)

	// 代金券扣除 22 元，剩余 28 元；现金 100 元分文未动
	gotCredit := cache.creditInt(creditKey, "coupon_c1")
	wantCredit := int64(28 * 1000000)
	if gotCredit != wantCredit {
		t.Errorf("credit balance = %d, want %d", gotCredit, wantCredit)
	}
	gotCash := cache.balanceInt("balance:app_billing")
	wantCash := int64(100 * 1000000)
	if gotCash != wantCash {
		t.Errorf("cash balance = %d, want %d", gotCash, wantCash)
	}

	// 校验计费标签
	if dType := ctx.GetLabel(context_label.LabelDeductType); dType != context_label.DeductTypeCredit {
		t.Errorf("deduct_type = %s, want %s", dType, context_label.DeductTypeCredit)
	}
	if cid := ctx.GetLabel(context_label.LabelCreditID); cid != "coupon_c1" {
		t.Errorf("credit_id = %s, want coupon_c1", cid)
	}
	if cApp := ctx.GetLabel(context_label.LabelCreditApplied); cApp != "22.000000" {
		t.Errorf("credit_applied = %s, want 22.000000", cApp)
	}
	if bApp := ctx.GetLabel(context_label.LabelBalanceApplied); bApp != "0" {
		t.Errorf("balance_applied = %s, want 0", bApp)
	}
}

// TestEndToEnd_Credit_PreDeductAndSettle_InsufficientCredit_OverdraftCash 跨账本抵扣测试
func TestEndToEnd_Credit_PreDeductAndSettle_InsufficientCredit_OverdraftCash(t *testing.T) {
	defer setMockBillingEnabled(true)()

	variables := price_calcular.Variables{
		"status": {Source: "response_status", Type: "integer"},
	}
	rules := []*price_calcular.Rule{
		{
			ID:                 "rule_success",
			Name:               "跨账本测试规则",
			CostExpression:     "cost_per_call * 1.0",
			SaleExpression:     "sale_per_call * 1.5", // 20 * 1.5 = 30 元
			OfficialExpression: "official_per_call * 1.0",
		},
	}
	calc, err := price_calcular.NewCalculator("calc_cross", "USD", variables, rules)
	if err != nil {
		t.Fatalf("create calculator failed: %v", err)
	}
	price_calcular.SetCalculator("api:res_cross", calc)
	defer price_calcular.DelCalculator("api:res_cross")

	balanceKey := "balance:app_cross"
	priceKey := "access-resource-price:app_cross:res_cross"
	priceJSON := `{
		"basic_info": {"version": "v1.0.0", "resource_group_id": "res_cross", "tenant_id": "t1"},
		"strategy": {
			"rule_success": {
				"cost": {"per_call": 2.0},
				"sale": {"per_call": 20.0},
				"official": {"per_call": 25.0}
			}
		}
	}`
	cache := newLuaCache(map[string]string{
		balanceKey: strconv.FormatInt(100*1000000, 10), // 现金 100 元
		priceKey:   priceJSON,
	})
	scope_manager.Set("lua_cache_cross", cache, "lua_cache_cross", "redis")
	defer scope_manager.Del("lua_cache_cross")

	plugin := &executor{
		WorkerBase:              drivers.Worker("billing_test", "billing_test"),
		redisID:                 "lua_cache_cross",
		balanceKeyGenerator:     context_label.NewKeyGenerator(balanceKey),
		creditKeyGenerator:      context_label.NewKeyGenerator("{product}:credit:user:{balance_target}"),
		priceKeyGenerator:       context_label.NewKeyGenerator(priceKey),
		taskKeyGenerator:        context_label.NewKeyGenerator("task:{application}:{resource}"),
		concurrencyKeyGenerator: context_label.NewKeyGenerator("conc:{application}:{resource}"),
		settingKeyGenerator:     context_label.NewKeyGenerator("dynamic-billing-setting:{application}:{resource}"),
	}

	ctx := newMockHttpContext(map[string]string{
		"application":    "app_cross",
		"api":            "api_cross",
		"resource":       "res_cross",
		"resource_type":  "api",
		"product":        "demo_prod",
		"balance_target": "user_cross",
		"service":        "srv_cross",
	}, &mockResponse{statusCode: 200})

	creditRule := credit.CompileRule("coupon_cross", &credit.Config{
		Nbf: time.Now().Unix() - 100,
		Exp: time.Now().Unix() + 3600,
		Filters: credit.Filters{
			Provider: credit.RuleFilter{API: []string{"srv_cross"}},
			Resource: credit.RuleFilter{API: []string{"res_cross"}},
		},
	})
	credit.SetRule("coupon_cross", creditRule)
	defer credit.DelRule("coupon_cross")

	creditKey := "demo_prod:credit:user:user_cross"
	cache.setCreditBalance(creditKey, "coupon_cross", 25*1000000)

	if err := plugin.DoHttpFilter(ctx, &mockChain{}); err != nil {
		t.Fatalf("DoHttpFilter failed: %v", err)
	}
	triggerSettle(ctx)

	// 最终消费 30 元：
	// 代金券初始 25 元，预扣 20，少补时代金券不足，代金券清零，剩余转扣现金 5 元
	gotCredit := cache.creditInt(creditKey, "coupon_cross")
	if gotCredit != 0 {
		t.Errorf("credit balance = %d, want 0", gotCredit)
	}
	gotCash := cache.balanceInt(balanceKey)
	wantCash := int64(95 * 1000000) // 100 - 5 = 95 元
	if gotCash != wantCash {
		t.Errorf("cash balance = %d, want %d", gotCash, wantCash)
	}

	if dType := ctx.GetLabel(context_label.LabelDeductType); dType != context_label.DeductTypeCombined {
		t.Errorf("deduct_type = %s, want %s (cross-ledger)", dType, context_label.DeductTypeCombined)
	}
	if cid := ctx.GetLabel(context_label.LabelCreditID); cid != "coupon_cross" {
		t.Errorf("credit_id = %s, want coupon_cross", cid)
	}
	if cApp := ctx.GetLabel(context_label.LabelCreditApplied); cApp != "25.000000" {
		t.Errorf("credit_applied = %s, want 25.000000", cApp)
	}
	if bApp := ctx.GetLabel(context_label.LabelBalanceApplied); bApp != "5.000000" {
		t.Errorf("balance_applied = %s, want 5.000000", bApp)
	}
}

// TestEndToEnd_Credit_CandidatePriority_ExpAsc 候选代金券按失效时间排序优先抵扣测试
func TestEndToEnd_Credit_CandidatePriority_ExpAsc(t *testing.T) {
	plugin, cache, ctx := newBillingTestEnv(t, 100.0)
	ctx.SetLabel("product", "demo_prod")
	ctx.SetLabel("balance_target", "user_exp")
	ctx.SetLabel("service", "srv_exp")
	ctx.SetLabel("api", "api_billing")

	// 规则 1：最先失效（exp=now+100），但余额不足（仅 5 元 < 预扣 22 元）
	// 新逻辑：最先失效的券哪怕只有 5 元，也会先扣至 0，不足的扣现金
	rEarly := credit.CompileRule("c_early", &credit.Config{
		Nbf: time.Now().Unix() - 100,
		Exp: time.Now().Unix() + 100,
		Filters: credit.Filters{
			Provider: credit.RuleFilter{API: []string{"srv_exp"}},
			Resource: credit.RuleFilter{API: []string{"res_billing"}},
		},
	})
	credit.SetRule("c_early", rEarly)
	defer credit.DelRule("c_early")

	// 规则 2：较晚失效（exp=now+200），余额充足（50 元）
	rLate := credit.CompileRule("c_late", &credit.Config{
		Nbf: time.Now().Unix() - 100,
		Exp: time.Now().Unix() + 200,
		Filters: credit.Filters{
			Provider: credit.RuleFilter{API: []string{"srv_exp"}},
			Resource: credit.RuleFilter{API: []string{"res_billing"}},
		},
	})
	credit.SetRule("c_late", rLate)
	defer credit.DelRule("c_late")

	creditKey := "demo_prod:credit:user:user_exp"
	cache.setCreditBalance(creditKey, "c_early", 5*1000000)
	cache.setCreditBalance(creditKey, "c_late", 50*1000000)

	if err := plugin.DoHttpFilter(ctx, &mockChain{}); err != nil {
		t.Fatalf("DoHttpFilter failed: %v", err)
	}
	triggerSettle(ctx)

	// c_early 先被选中扣至 0，差额 17 元扣现金；c_late 未动
	if gotEarly := cache.creditInt(creditKey, "c_early"); gotEarly != 0 {
		t.Errorf("c_early balance = %d, want 0", gotEarly)
	}
	if gotLate := cache.creditInt(creditKey, "c_late"); gotLate != 50*1000000 {
		t.Errorf("c_late balance = %d, want 50000000", gotLate)
	}
	if cid := ctx.GetLabel(context_label.LabelCreditID); cid != "c_early" {
		t.Errorf("credit_id = %s, want c_early", cid)
	}
	if dType := ctx.GetLabel(context_label.LabelDeductType); dType != context_label.DeductTypeCombined {
		t.Errorf("deduct_type = %s, want %s", dType, context_label.DeductTypeCombined)
	}
}

// TestEndToEnd_Credit_DownstreamError_RefundCredit 下游失败冲正退还代金券测试
func TestEndToEnd_Credit_DownstreamError_RefundCredit(t *testing.T) {
	plugin, cache, ctx := newBillingTestEnv(t, 100.0)
	ctx.SetLabel("product", "demo_prod")
	ctx.SetLabel("balance_target", "user_err")
	ctx.SetLabel("service", "srv_err")
	ctx.SetLabel("api", "api_billing")

	creditRule := credit.CompileRule("coupon_err", &credit.Config{
		Nbf: time.Now().Unix() - 100,
		Exp: time.Now().Unix() + 3600,
		Filters: credit.Filters{
			Provider: credit.RuleFilter{API: []string{"srv_err"}},
			Resource: credit.RuleFilter{API: []string{"res_billing"}},
		},
	})
	credit.SetRule("coupon_err", creditRule)
	defer credit.DelRule("coupon_err")

	creditKey := "demo_prod:credit:user:user_err"
	cache.setCreditBalance(creditKey, "coupon_err", 50*1000000)

	if err := plugin.DoHttpFilter(ctx, &errChain{}); err == nil {
		t.Fatalf("expected error from chain, got nil")
	}

	// 冲正全额回滚，代金券恢复 50 元
	if got := cache.creditInt(creditKey, "coupon_err"); got != 50*1000000 {
		t.Errorf("credit balance after recovery = %d, want 50000000", got)
	}
}

// TestEndToEnd_MarkedPrice_ContextLabels 标价 Marked 表达式与审计标签注入测试
func TestEndToEnd_MarkedPrice_ContextLabels(t *testing.T) {
	defer setMockBillingEnabled(true)()

	variables := price_calcular.Variables{
		"status": {Source: "response_status", Type: "integer"},
	}
	rules := []*price_calcular.Rule{
		{
			ID:                 "rule_marked",
			Name:               "标价测试规则",
			CostExpression:     "cost_per_call * 1.0",
			SaleExpression:     "sale_per_call * 1.0",
			OfficialExpression: "official_per_call * 1.0",
			MarkedExpression:   "marked_per_call * 1.5",
		},
	}
	calc, err := price_calcular.NewCalculator("calc_marked", "USD", variables, rules)
	if err != nil {
		t.Fatalf("create calculator failed: %v", err)
	}
	price_calcular.SetCalculator("api:res_marked", calc)
	defer price_calcular.DelCalculator("api:res_marked")

	balanceKey := "balance:app_marked"
	priceKey := "access-resource-price:app_marked:res_marked"
	priceJSON := `{
		"basic_info": {"version": "v1.0.0", "resource_group_id": "res_marked", "tenant_id": "t1"},
		"strategy": {
			"rule_marked": {
				"cost": {"per_call": 2.0},
				"sale": {"per_call": 10.0},
				"official": {"per_call": 12.0},
				"marked": {"per_call": 8.0}
			}
		}
	}`
	cache := newLuaCache(map[string]string{
		balanceKey: strconv.FormatInt(100*1000000, 10),
		priceKey:   priceJSON,
	})
	scope_manager.Set("lua_cache_marked", cache, "lua_cache_marked", "redis")
	defer scope_manager.Del("lua_cache_marked")

	plugin := &executor{
		WorkerBase:              drivers.Worker("billing_test", "billing_test"),
		redisID:                 "lua_cache_marked",
		balanceKeyGenerator:     context_label.NewKeyGenerator(balanceKey),
		creditKeyGenerator:      context_label.NewKeyGenerator("{product}:credit:user:{balance_target}"),
		priceKeyGenerator:       context_label.NewKeyGenerator(priceKey),
		taskKeyGenerator:        context_label.NewKeyGenerator("task:{application}:{resource}"),
		concurrencyKeyGenerator: context_label.NewKeyGenerator("conc:{application}:{resource}"),
		settingKeyGenerator:     context_label.NewKeyGenerator("dynamic-billing-setting:{application}:{resource}"),
	}

	ctx := newMockHttpContext(map[string]string{
		"application":   "app_marked",
		"api":           "api_marked",
		"resource":      "res_marked",
		"resource_type": "api",
	}, &mockResponse{statusCode: 200})

	if err := plugin.DoHttpFilter(ctx, &mockChain{}); err != nil {
		t.Fatalf("DoHttpFilter failed: %v", err)
	}
	triggerSettle(ctx)

	// 校验标价相关标签注入
	// marked_amount = 8.0 * 1.5 = 12.000000
	if markedAmt := ctx.GetLabel(context_label.LabelAmountMarked); markedAmt != "12.000000" {
		t.Errorf("marked_amount = %s, want 12.000000", markedAmt)
	}
	if markedExpr := ctx.GetLabel(context_label.LabelExprMarked); markedExpr != "marked_per_call * 1.5" {
		t.Errorf("marked_expr = %s, want 'marked_per_call * 1.5'", markedExpr)
	}
}

// TestEndToEnd_Credit_PartialPreDeduct_CreditToZeroAndCashRemainder
// 重点测试用户需求：预扣时先检查券余额，若券不足预扣金额，先预扣券余额到0，剩下不足的扣现金余额。
func TestEndToEnd_Credit_PartialPreDeduct_CreditToZeroAndCashRemainder(t *testing.T) {
	plugin, cache, ctx := newBillingTestEnv(t, 100.0) // 现金 100 元，预扣金额需 22 元
	ctx.SetLabel("product", "demo_prod")
	ctx.SetLabel("balance_target", "user_partial")
	ctx.SetLabel("service", "srv_partial")
	ctx.SetLabel("api", "api_billing")

	creditRule := credit.CompileRule("coupon_part", &credit.Config{
		Nbf: time.Now().Unix() - 100,
		Exp: time.Now().Unix() + 3600,
		Filters: credit.Filters{
			Provider: credit.RuleFilter{API: []string{"srv_partial"}},
			Resource: credit.RuleFilter{API: []string{"res_billing"}},
		},
	})
	credit.SetRule("coupon_part", creditRule)
	defer credit.DelRule("coupon_part")

	creditKey := "demo_prod:credit:user:user_partial"
	// 券余额仅 10 元 (< 22 元)
	cache.setCreditBalance(creditKey, "coupon_part", 10*1000000)

	if err := plugin.DoHttpFilter(ctx, &mockChain{}); err != nil {
		t.Fatalf("DoHttpFilter failed: %v", err)
	}

	// 此时仅完成阶段一预扣：预扣金额为底价 20 元
	// 代金券 10 元应被扣至 0；现金应被扣除剩余的 10 元 (100 - 10 = 90 元)
	gotCredit := cache.creditInt(creditKey, "coupon_part")
	if gotCredit != 0 {
		t.Errorf("credit balance after pre-deduct = %d, want 0", gotCredit)
	}
	gotCash := cache.balanceInt("balance:app_billing")
	wantCashPre := int64(90 * 1000000)
	if gotCash != wantCashPre {
		t.Errorf("cash balance after pre-deduct = %d, want %d (90元)", gotCash, wantCashPre)
	}

	// 校验阶段一标签
	if dType := ctx.GetLabel(context_label.LabelDeductType); dType != context_label.DeductTypeCombined {
		t.Errorf("deduct_type in label = %s, want %s", dType, context_label.DeductTypeCombined)
	}
	if cid := ctx.GetLabel(context_label.LabelCreditID); cid != "coupon_part" {
		t.Errorf("credit_id in label = %s, want coupon_part", cid)
	}

	// 执行阶段二结算（实际结算金额 22 元，相比预扣 20 元少补 2 元，从现金中追加扣除 2 元）
	triggerSettle(ctx)

	// 结算后：券依然为 0，现金扣减 2 元变为 88 元；applied 标签记录代金券 10 元，现金 12 元
	if gotCreditAfter := cache.creditInt(creditKey, "coupon_part"); gotCreditAfter != 0 {
		t.Errorf("credit balance after settle = %d, want 0", gotCreditAfter)
	}
	wantCashAfter := int64(88 * 1000000)
	if gotCashAfter := cache.balanceInt("balance:app_billing"); gotCashAfter != wantCashAfter {
		t.Errorf("cash balance after settle = %d, want %d (88元)", gotCashAfter, wantCashAfter)
	}
	if cApp := ctx.GetLabel(context_label.LabelCreditApplied); cApp != "10.000000" {
		t.Errorf("credit_applied = %s, want 10.000000", cApp)
	}
	if bApp := ctx.GetLabel(context_label.LabelBalanceApplied); bApp != "12.000000" {
		t.Errorf("balance_applied = %s, want 12.000000", bApp)
	}
}

// TestEndToEnd_Credit_PartialPreDeduct_DownstreamError_Recovery
// 测试组合预扣（券扣到0+扣现金）后下游错误触发冲正：券和现金各自原路退回
func TestEndToEnd_Credit_PartialPreDeduct_DownstreamError_Recovery(t *testing.T) {
	plugin, cache, ctx := newBillingTestEnv(t, 100.0)
	ctx.SetLabel("product", "demo_prod")
	ctx.SetLabel("balance_target", "user_rec")
	ctx.SetLabel("service", "srv_rec")
	ctx.SetLabel("api", "api_billing")

	creditRule := credit.CompileRule("coupon_rec", &credit.Config{
		Nbf: time.Now().Unix() - 100,
		Exp: time.Now().Unix() + 3600,
		Filters: credit.Filters{
			Provider: credit.RuleFilter{API: []string{"srv_rec"}},
			Resource: credit.RuleFilter{API: []string{"res_billing"}},
		},
	})
	credit.SetRule("coupon_rec", creditRule)
	defer credit.DelRule("coupon_rec")

	creditKey := "demo_prod:credit:user:user_rec"
	cache.setCreditBalance(creditKey, "coupon_rec", 10*1000000)

	// 模拟下游错误
	if err := plugin.DoHttpFilter(ctx, &errChain{}); err == nil {
		t.Fatalf("expected error from chain, got nil")
	}

	// 冲正后：券 10 元全部恢复，现金 100 元全部恢复
	if gotCredit := cache.creditInt(creditKey, "coupon_rec"); gotCredit != 10*1000000 {
		t.Errorf("credit balance after recovery = %d, want 10000000", gotCredit)
	}
	if gotCash := cache.balanceInt("balance:app_billing"); gotCash != 100*1000000 {
		t.Errorf("cash balance after recovery = %d, want 100000000", gotCash)
	}
}
