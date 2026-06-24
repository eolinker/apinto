package dynamic_billing

import (
	"context"
	"encoding/json"
	"github.com/eolinker/apinto/utils/context-label"
	"math"
	"net/http"
	"testing"
	"time"

	"github.com/eolinker/apinto/drivers"
	pricing_policy "github.com/eolinker/apinto/drivers/pricing-policy"
	"github.com/eolinker/apinto/drivers/pricing-policy/manager"
	"github.com/eolinker/apinto/resources"
	scope_manager "github.com/eolinker/apinto/scope-manager"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
)

// mockStringResult 模拟 ICache 的返回结果
type mockStringResult struct {
	val string
	err error
}

func (m *mockStringResult) Result() (string, error) {
	return m.val, m.err
}

func (m *mockStringResult) Bytes() ([]byte, error) {
	return []byte(m.val), m.err
}

// mockIntResult 模拟 IntResult 返回结果
type mockIntResult struct {
	val int64
	err error
}

func (m *mockIntResult) Result() (int64, error) {
	return m.val, m.err
}

// mockInterfaceResult 模拟 InterfaceResult 返回结果
type mockInterfaceResult struct {
	val interface{}
	err error
}

func (m *mockInterfaceResult) Result() (interface{}, error) {
	return m.val, m.err
}

// mockBoolResult 模拟 BoolResult 返回结果
type mockBoolResult struct {
	val bool
	err error
}

func (m *mockBoolResult) Result() (bool, error) {
	return m.val, m.err
}

// mockCache 模拟 Redis 缓存资源
type mockCache struct {
	resources.ICache
	val string
	err error
	kv  map[string]string
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
	if m.kv == nil {
		m.kv = make(map[string]string)
	}
	m.kv[key] = string(value)
	return nil
}

func (m *mockCache) SetNX(ctx context.Context, key string, value []byte, expiration time.Duration) resources.BoolResult {
	if m.kv == nil {
		m.kv = make(map[string]string)
	}
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
	return &mockInterfaceResult{val: int64(1), err: nil}
}

func (m *mockCache) Id() string                                               { return "mock_cache" }
func (m *mockCache) Name() string                                             { return "mock_cache" }
func (m *mockCache) Start() error                                             { return nil }
func (m *mockCache) Reset(interface{}, map[eosc.RequireId]eosc.IWorker) error { return nil }
func (m *mockCache) Stop() error                                              { return nil }
func (m *mockCache) CheckSkill(string) bool                                   { return true }

// mockManager 模拟 pricing-policy 的 manager.IManager
type mockManager struct {
	manager.IManager
	executors map[string]manager.IPolicyExecutor
}

func (m *mockManager) Get(id string) (manager.IPolicyExecutor, bool) {
	e, ok := m.executors[id]
	return e, ok
}

// mockPolicyExecutor 模拟 pricing-policy 驱动，并实现 manager.IPolicyExecutor
type mockPolicyExecutor struct {
	eosc.IWorker
	calc *pricing_policy.Calculator
}

func (m *mockPolicyExecutor) Calculator() interface{} {
	return m.calc
}

func (m *mockPolicyExecutor) Id() string               { return "res_01" }
func (m *mockPolicyExecutor) Name() string             { return "res_01" }
func (m *mockPolicyExecutor) CheckSkill(s string) bool { return true }

// mockResponse 模拟 HTTP 响应
type mockResponse struct {
	http_context.IResponse
	statusCode int
}

func (m *mockResponse) StatusCode() int {
	return m.statusCode
}

func (m *mockResponse) ContentLength() int {
	return 0
}

func (m *mockResponse) ContentEncoding() []byte {
	return nil
}

func (m *mockResponse) GetBody() []byte {
	return nil
}

func (m *mockResponse) IsBodyStream() bool {
	return false
}

// mockHttpContext 模拟 http_context.IHttpContext
type mockHttpContext struct {
	http_context.IHttpContext
	labels map[string]string
	values map[string]interface{}
	ctx    context.Context
	resp   *mockResponse
}

func (m *mockHttpContext) GetLabel(key string) string {
	return m.labels[key]
}

func (m *mockHttpContext) Assert(i interface{}) error {
	if v, ok := i.(*http_context.IHttpContext); ok {
		*v = m
		return nil
	}
	return nil
}

func (m *mockHttpContext) SetLabel(key, val string) {
	m.labels[key] = val
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

func (m *mockHttpContext) Context() context.Context {
	return m.ctx
}

func (m *mockHttpContext) Response() http_context.IResponse {
	return m.resp
}

func (m *mockHttpContext) Request() http_context.IRequestReader {
	return nil
}

// mockChain 模拟调用链
type mockChain struct {
	eocontext.IChain
}

func (m *mockChain) DoChain(ctx eocontext.EoContext) error {
	return nil
}

func TestDynamicBilling_Executor(t *testing.T) {
	// 1. 初始化 Mock pricing-policy manager 并注入全局变量
	mockMgr := &mockManager{
		executors: make(map[string]manager.IPolicyExecutor),
	}
	policyManager = mockMgr

	// 2. 创建真实定价计算器 pricing-policy calculator
	policyCfg := &pricing_policy.Config{
		Currency: "USD",
		ContextVariables: map[string]*pricing_policy.Variable{
			"status": {
				Source: "response_status",
				Type:   "integer",
			},
		},
		AdvancedRules: []*pricing_policy.Rule{
			{
				ID:   "rule_success",
				Name: "成功请求规则",
				Conditions: &pricing_policy.Condition{
					AllOf: []*pricing_policy.AllOf{
						{
							BasicRule: &pricing_policy.BasicRule{
								Key:   "status",
								Op:    "==",
								Value: "200",
								Type:  "integer",
							},
						},
					},
				},
				CostExpression:     "cost_per_call * 1.0",
				SaleExpression:     "sale_per_call * 1.1",
				OfficialExpression: "official_per_call * 1.2",
			},
		},
	}

	calc, err := pricing_policy.NewCalculator(policyCfg)
	if err != nil {
		t.Fatalf("create calculator failed: %v", err)
	}

	mockMgr.executors["res_01"] = &mockPolicyExecutor{
		calc: calc,
	}

	// 3. 构建 mock Redis 数据
	redisPriceJSON := `{
		"basic_info": {
			"version": "v1.0.0",
			"rely": "none",
			"resource_group_id": "res_01",
			"tenant_id": "tenant_01"
		},
		"strategy": {
			"base": {
				"cost": { "per_call": 1.0 },
				"sale": { "per_call": 10.0 },
				"official": { "per_call": 12.0 }
			},
			"rule_success": {
				"cost": { "per_call": 2.0 },
				"sale": { "per_call": 20.0 },
				"official": { "per_call": 25.0 }
			}
		}
	}`

	cache := &mockCache{
		val: redisPriceJSON,
	}

	// 将 mockCache 注册进 scope_manager
	scope_manager.Set("mock_cache", cache, "redis")
	defer scope_manager.Del("mock_cache")

	// 4. 创建 resource-pricing plugin 实例
	plugin := &executor{
		WorkerBase:              drivers.Worker("pricing_filter_test", "pricing_filter_test"),
		redisID:                 "mock_cache",
		balanceKeyGenerator:     context_label.NewKeyGenerator("balance:{application}"),
		priceKeyGenerator:       context_label.NewKeyGenerator("access-resource-price:{application}:{resource}"),
		taskKeyGenerator:        context_label.NewKeyGenerator("resource-pricing-task:{application}:{resource}"),
		concurrencyKeyGenerator: context_label.NewKeyGenerator("resource-pricing-concurrency:{application}:{resource}"),
	}

	// 5. 模拟一次命中 200 成功的 HTTP 请求
	ctx := &mockHttpContext{
		labels: map[string]string{
			"resource_id": "res_01",
			"application": "app_client_01",
		},
		values: make(map[string]interface{}),
		ctx:    context.Background(),
		resp: &mockResponse{
			statusCode: 200,
		},
	}

	chain := &mockChain{}

	// 6. 执行插件过滤
	err = plugin.DoHttpFilter(ctx, chain)
	if err != nil {
		t.Fatalf("DoHttpFilter failed: %v", err)
	}

	// 7. 验证断言结果
	if ctx.labels["pricing_status"] != "success" {
		t.Errorf("expected pricing_status success, got %v", ctx.labels["pricing_status"])
	}

	// 命中 rule_success 规则，该规则价格:
	// cost_per_call = 2.0 => Cost 表达式: cost_per_call * 1.0 = 2.0
	// sale_per_call = 20.0 => Sale 表达式: sale_per_call * 1.1 = 22.0
	// official_per_call = 25.0 => Official 表达式: official_per_call * 1.2 = 30.0
	expectedCost := 2.0
	expectedSale := 22.0
	expectedOfficial := 30.0

	costVal, ok := ctx.values["pricing_cost"].(float64)
	if !ok || math.Abs(costVal-expectedCost) > 1e-9 {
		t.Errorf("expected cost %v, got %v", expectedCost, costVal)
	}

	saleVal, ok := ctx.values["pricing_sale"].(float64)
	if !ok || math.Abs(saleVal-expectedSale) > 1e-9 {
		t.Errorf("expected sale %v, got %v", expectedSale, saleVal)
	}

	officialVal, ok := ctx.values["pricing_official"].(float64)
	if !ok || math.Abs(officialVal-expectedOfficial) > 1e-9 {
		t.Errorf("expected official %v, got %v", expectedOfficial, officialVal)
	}

	if ctx.labels["pricing_currency"] != "USD" {
		t.Errorf("expected pricing_currency USD, got %v", ctx.labels["pricing_currency"])
	}
}

// 保证所有依赖 of mockResponse 接口都被实现
func (m *mockResponse) Headers() http.Header        { return nil }
func (m *mockResponse) ResponseTime() time.Duration { return 0 }
func (m *mockResponse) ResponseError() error        { return nil }

func TestDynamicBilling_ConcurrencyAndBalance(t *testing.T) {
	// 1. 初始化 Mock pricing-policy manager 并注入全局变量
	mockMgr := &mockManager{
		executors: make(map[string]manager.IPolicyExecutor),
	}
	policyManager = mockMgr

	// 2. 创建真实定价计算器 pricing-policy calculator
	policyCfg := &pricing_policy.Config{
		Currency: "USD",
		ContextVariables: map[string]*pricing_policy.Variable{
			"status": {
				Source: "response_status",
				Type:   "integer",
			},
		},
		AdvancedRules: []*pricing_policy.Rule{
			{
				ID:   "rule_success",
				Name: "成功请求规则",
				Conditions: &pricing_policy.Condition{
					AllOf: []*pricing_policy.AllOf{
						{
							BasicRule: &pricing_policy.BasicRule{
								Key:   "status",
								Op:    "==",
								Value: "200",
								Type:  "integer",
							},
						},
					},
				},
				CostExpression:     "cost_per_call * 1.0",
				SaleExpression:     "sale_per_call * 1.1",
				OfficialExpression: "official_per_call * 1.2",
			},
		},
	}

	calc, err := pricing_policy.NewCalculator(policyCfg)
	if err != nil {
		t.Fatalf("create calculator failed: %v", err)
	}

	mockMgr.executors["res_01"] = &mockPolicyExecutor{
		calc: calc,
	}

	// 3. 构建 mock Redis 数据
	redisPriceJSON := `{
		"basic_info": {
			"version": "v1.0.0",
			"rely": "none",
			"resource_group_id": "res_01",
			"tenant_id": "tenant_01"
		},
		"strategy": {
			"base": {
				"cost": { "per_call": 1.0 },
				"sale": { "per_call": 10.0 },
				"official": { "per_call": 12.0 }
			}
		}
	}`

	cache := &mockCache{
		val: redisPriceJSON,
	}

	// 将 mockCache 注册进 scope_manager
	scope_manager.Set("mock_cache", cache, "redis")
	defer scope_manager.Del("mock_cache")

	// 4. 创建 resource-pricing plugin 实例 (启用余额扣减，并限制并发为 1)
	plugin := &executor{
		WorkerBase:              drivers.Worker("pricing_filter_test", "pricing_filter_test"),
		redisID:                 "mock_cache",
		defaultConcurrencyLimit: 1,
		enableBalance:           true,
		balanceKeyGenerator:     context_label.NewKeyGenerator("balance:{user}"),
		priceKeyGenerator:       context_label.NewKeyGenerator("access-resource-price:{user}:{resource}"),
		taskKeyGenerator:        context_label.NewKeyGenerator("resource-pricing-task:{user}:{resource}"),
		concurrencyKeyGenerator: context_label.NewKeyGenerator("resource-pricing-concurrency:{user}:{resource}"),
	}

	// 5. 模拟一次 HTTP 请求 (余额模拟为 100)
	ctx := &mockHttpContext{
		labels: map[string]string{
			"resource_id": "res_01",
			"user":        "user_01",
		},
		values: make(map[string]interface{}),
		ctx:    context.Background(),
		resp: &mockResponse{
			statusCode: 200,
		},
	}

	chain := &mockChain{}

	err = plugin.DoHttpFilter(ctx, chain)
	if err != nil {
		t.Fatalf("DoHttpFilter failed: %v", err)
	}

	if ctx.labels["pricing_status"] != "success" {
		t.Errorf("expected pricing_status success, got %v", ctx.labels["pricing_status"])
	}
}

func TestDynamicBilling_TaskPricingSnapshot(t *testing.T) {
	// 1. 初始化计费策略 Worker
	mockMgr := &mockManager{
		executors: make(map[string]manager.IPolicyExecutor),
	}
	policyManager = mockMgr

	// 2. 创建真实定价计算器 pricing-policy calculator
	policyCfg := &pricing_policy.Config{
		Currency: "USD",
		ContextVariables: map[string]*pricing_policy.Variable{
			"status": {
				Source: "response_status",
				Type:   "integer",
			},
		},
		AdvancedRules: []*pricing_policy.Rule{
			{
				ID:   "rule_success",
				Name: "成功请求规则",
				Conditions: &pricing_policy.Condition{
					AllOf: []*pricing_policy.AllOf{
						{
							BasicRule: &pricing_policy.BasicRule{
								Key:   "status",
								Op:    "==",
								Value: "200",
								Type:  "integer",
							},
						},
					},
				},
				CostExpression:     "cost_per_call",
				SaleExpression:     "sale_per_call",
				OfficialExpression: "official_per_call",
			},
		},
	}

	calc, err := pricing_policy.NewCalculator(policyCfg)
	if err != nil {
		t.Fatalf("create calculator failed: %v", err)
	}

	mockMgr.executors["res_snapshot"] = &mockPolicyExecutor{
		calc: calc,
	}

	// 3. 构建 mock Redis 数据 (初始价格是 10.0 每次)
	redisPriceJSON := `{
		"basic_info": {
			"version": "v1.0.0",
			"resource_group_id": "res_snapshot",
			"tenant_id": "tenant_01"
		},
		"strategy": {
			"rule_success": {
				"cost": { "per_call": 10.0 },
				"sale": { "per_call": 10.0 },
				"official": { "per_call": 10.0 }
			}
		}
	}`

	cache := &mockCache{
		val: redisPriceJSON,
		kv:  make(map[string]string),
	}

	// 注入初始的价格规则
	priceKey := "access-resource-price:user_snapshot:res_snapshot"
	cache.kv[priceKey] = redisPriceJSON

	// 注册 cache
	scope_manager.Set("mock_cache_snapshot", cache, "redis")
	defer scope_manager.Del("mock_cache_snapshot")

	// 4. 创建 dynamic-billing plugin 实例
	plugin := &executor{
		WorkerBase:              drivers.Worker("pricing_filter_test", "pricing_filter_test"),
		redisID:                 "mock_cache_snapshot",
		defaultConcurrencyLimit: 100,
		enableBalance:           true,
		balanceKeyGenerator:     context_label.NewKeyGenerator("balance:{user}"),
		priceKeyGenerator:       context_label.NewKeyGenerator("access-resource-price:{user}:{resource}"),
		taskKeyGenerator:        context_label.NewKeyGenerator("dynamic-billing-task:{user}:{resource}"),
		concurrencyKeyGenerator: context_label.NewKeyGenerator("dynamic-billing-concurrency:{user}:{resource}"),
	}

	// 5. 模拟 Task Create 请求
	ctxCreate := &mockHttpContext{
		labels: map[string]string{
			"resource_id":  "res_snapshot",
			"user":         "user_snapshot",
			"task_id":      "task_abc_123",
			"billing_mode": "task_create",
		},
		values: make(map[string]interface{}),
		ctx:    context.Background(),
		resp: &mockResponse{
			statusCode: 200,
		},
	}

	chain := &mockChain{}
	err = plugin.DoHttpFilter(ctxCreate, chain)
	if err != nil {
		t.Fatalf("TaskCreate DoHttpFilter failed: %v", err)
	}

	// 验证 TaskInfo 已经在 Redis 里固化了
	taskKey := "dynamic-billing-task:user_snapshot:res_snapshot"
	taskInfoStr, hasTask := cache.kv[taskKey]
	if !hasTask {
		t.Fatalf("expected TaskInfo in redis, but not found")
	}

	var taskInfo TaskInfo
	err = json.Unmarshal([]byte(taskInfoStr), &taskInfo)
	if err != nil {
		t.Fatalf("unmarshal TaskInfo failed: %v", err)
	}
	if taskInfo.PricingData == nil {
		t.Fatalf("expected pricing data snapshotted in TaskInfo, but got nil")
	}

	// 6. 篡改 Redis 中的实时价格规则 (篡改为 50.0 每次)
	modifiedPriceJSON := `{
		"basic_info": {
			"version": "v1.0.0",
			"resource_group_id": "res_snapshot",
			"tenant_id": "tenant_01"
		},
		"strategy": {
			"rule_success": {
				"cost": { "per_call": 50.0 },
				"sale": { "per_call": 50.0 },
				"official": { "per_call": 50.0 }
			}
		}
	}`
	cache.kv[priceKey] = modifiedPriceJSON

	// 7. 模拟 Task Query 请求 (应当遵循快照里的 10.0 价格，而不是篡改后的 50.0)
	ctxQuery := &mockHttpContext{
		labels: map[string]string{
			"resource_id":  "res_snapshot",
			"user":         "user_snapshot",
			"task_id":      "task_abc_123",
			"billing_mode": "task_query",
		},
		values: make(map[string]interface{}),
		ctx:    context.Background(),
		resp: &mockResponse{
			statusCode: 200,
		},
	}

	err = plugin.DoHttpFilter(ctxQuery, chain)
	if err != nil {
		t.Fatalf("TaskQuery DoHttpFilter failed: %v", err)
	}

	costStr := ctxQuery.labels["pricing_cost"]
	if costStr != "10.000000" {
		t.Errorf("expected pricing_cost to be '10.000000' (from snapshotted pricing), but got '%s'", costStr)
	}
}
