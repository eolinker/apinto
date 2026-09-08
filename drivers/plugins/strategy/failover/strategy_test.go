package failover

import (
	"errors"
	"net"
	"testing"

	ai_convert "github.com/eolinker/apinto/ai-convert"
	"github.com/eolinker/apinto/drivers"
	failover_strategy "github.com/eolinker/apinto/drivers/strategy/failover-strategy"
	http_context "github.com/eolinker/apinto/node/http-context"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/valyala/fasthttp"
)

func newMockHttpContext() http_service.IHttpContext {
	fast := &fasthttp.RequestCtx{}
	freq := fasthttp.AcquireRequest()
	addr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:8080")
	fast.Init(freq, addr, nil)
	return http_context.NewContext(fast, 0)
}

type mockChain struct {
	fn func(ctx eocontext.EoContext) error
}

func (m *mockChain) DoChain(ctx eocontext.EoContext) error {
	if m.fn != nil {
		return m.fn(ctx)
	}
	return nil
}

func (m *mockChain) Destroy() {}

// mockConverterDriver 实现 ai_convert.IConverterDriver
type mockConverterDriver struct {
	provider          string
	modelType         ai_convert.ModelType
	requestConvertFn  func(ctx eocontext.EoContext, extender map[string]interface{}) error
	responseConvertFn func(ctx eocontext.EoContext) error
}

func (m *mockConverterDriver) Provider() string                 { return m.provider }
func (m *mockConverterDriver) ModelType() ai_convert.ModelType { return m.modelType }
func (m *mockConverterDriver) RequestConvert(ctx eocontext.EoContext, extender map[string]interface{}) error {
	if m.requestConvertFn != nil {
		return m.requestConvertFn(ctx, extender)
	}
	return nil
}

func (m *mockConverterDriver) ResponseConvert(ctx eocontext.EoContext) error {
	if m.responseConvertFn != nil {
		return m.responseConvertFn(ctx)
	}
	return nil
}

// mockKeyResource 实现 ai_convert.IKeyResource
type mockKeyResource struct {
	id        string
	priority  int
	health    bool
	isBreaker bool
	driver    ai_convert.IConverterDriver
}

func (m *mockKeyResource) ID() string        { return m.id }
func (m *mockKeyResource) Health() bool      { return m.health }
func (m *mockKeyResource) Priority() int     { return m.priority }
func (m *mockKeyResource) Up()               { m.health = true; m.isBreaker = false }
func (m *mockKeyResource) Down()             { m.health = false }
func (m *mockKeyResource) IsBreaker() bool   { return m.isBreaker }
func (m *mockKeyResource) Breaker()          { m.isBreaker = true }
func (m *mockKeyResource) Get(modelType ai_convert.ModelType) (ai_convert.IConverterDriver, bool) {
	return m.driver, true
}
func (m *mockKeyResource) ModelTypeList() []ai_convert.ModelType {
	return []ai_convert.ModelType{ai_convert.ModelTypeOpenAIChat}
}

// mockProvider 实现 ai_convert.IProvider
type mockProvider struct {
	id             string
	provider       string
	model          string
	priority       int
	health         bool
	modelConfig    map[string]interface{}
	balanceHandler eocontext.BalanceHandler
}

func (m *mockProvider) ID() string                                { return m.id }
func (m *mockProvider) Provider() string                          { return m.provider }
func (m *mockProvider) Model() string                             { return m.model }
func (m *mockProvider) ModelConfig() map[string]interface{}       { return m.modelConfig }
func (m *mockProvider) Priority() int                             { return m.priority }
func (m *mockProvider) Health() bool                              { return m.health }
func (m *mockProvider) Down()                                     { m.health = false }
func (m *mockProvider) BalanceHandler() eocontext.BalanceHandler { return m.balanceHandler }
func (m *mockProvider) GenExtender(cfg string) (map[string]interface{}, error) {
	return m.modelConfig, nil
}

func TestStrategy_NoMatch(t *testing.T) {
	strat := &Strategy{
		WorkerBase: drivers.Worker("failover-plugin-id", "failover-plugin"),
		modelType:  ai_convert.ModelTypeOpenAIChat,
	}

	ctx := newMockHttpContext()
	ctx.SetLabel("provider", "non-existent")

	chainCalled := false
	chain := &mockChain{
		fn: func(c eocontext.EoContext) error {
			chainCalled = true
			return nil
		},
	}

	err := strat.DoHttpFilter(ctx, chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !chainCalled {
		t.Fatalf("expected next.DoChain to be called when no strategy matches")
	}
}

func TestStrategy_ResourcePriorityOverProvider(t *testing.T) {
	// 注册基于 provider 的宽泛策略
	providerHandler, err := failover_strategy.NewFailoverHandler(&failover_strategy.Config{
		Name:        "strategy-provider",
		Priority:    1,
		TriggerType: failover_strategy.TriggerTypeDirect,
		Filters: map[string][]string{
			"provider": {"openai"},
		},
		Providers: []*failover_strategy.ProviderConf{
			{Name: "fallback-provider-1", Model: "backup-model-1"},
		},
	})
	if err != nil {
		t.Fatalf("create provider handler failed: %v", err)
	}

	// 注册基于 resource 的精准策略
	resourceHandler, err := failover_strategy.NewFailoverHandler(&failover_strategy.Config{
		Name:        "strategy-resource",
		Priority:    10,
		TriggerType: failover_strategy.TriggerTypeDirect,
		Filters: map[string][]string{
			"resource": {"openai/gpt-4o"},
		},
		Providers: []*failover_strategy.ProviderConf{
			{Name: "fallback-resource-1", Model: "backup-model-res"},
		},
	})
	if err != nil {
		t.Fatalf("create resource handler failed: %v", err)
	}

	extractor := failover_strategy.NewExtractor()
	extractor.Set("p1", providerHandler)
	extractor.Set("r1", resourceHandler)

	// 场景 1: 请求 resource 为 openai/gpt-4o 时，优先命中 resource 策略，忽略 provider 策略
	ctx1 := newMockHttpContext()
	ctx1.SetLabel("provider", "openai")
	ctx1.SetLabel("model", "gpt-4o")

	matched1 := extractor.Extract(ctx1)
	if matched1 == nil {
		t.Fatalf("expected strategy to match")
	}
	if matched1.Name() != "strategy-resource" {
		t.Fatalf("expected strategy-resource to match, got %s", matched1.Name())
	}

	// 场景 2: 请求 resource 为 openai/gpt-3.5-turbo 时，未命中 resource 策略，降级命中 provider 策略
	ctx2 := newMockHttpContext()
	ctx2.SetLabel("provider", "openai")
	ctx2.SetLabel("model", "gpt-3.5-turbo")

	matched2 := extractor.Extract(ctx2)
	if matched2 == nil {
		t.Fatalf("expected provider strategy to match")
	}
	if matched2.Name() != "strategy-provider" {
		t.Fatalf("expected strategy-provider to match, got %s", matched2.Name())
	}
}

func TestStrategy_DirectFailover(t *testing.T) {
	strat := &Strategy{
		WorkerBase: drivers.Worker("failover-plugin-id", "failover-plugin"),
		modelType:  ai_convert.ModelTypeOpenAIChat,
	}

	backupProvider := "azure-openai"
	backupModel := "gpt-4o"

	// Mock KeyResource
	convertReqCalled := false
	convertRespCalled := false
	ai_convert.SetKeyResource(backupProvider, &mockKeyResource{
		id:       "key-1",
		priority: 1,
		health:   true,
		driver: &mockConverterDriver{
			provider:  backupProvider,
			modelType: ai_convert.ModelTypeOpenAIChat,
			requestConvertFn: func(ctx eocontext.EoContext, extender map[string]interface{}) error {
				convertReqCalled = true
				return nil
			},
			responseConvertFn: func(ctx eocontext.EoContext) error {
				convertRespCalled = true
				return nil
			},
		},
	})
	defer ai_convert.DelKeyResource(backupProvider, "key-1")

	handler, err := failover_strategy.NewFailoverHandler(&failover_strategy.Config{
		Name:        "direct-failover-strat",
		Priority:    1,
		TriggerType: failover_strategy.TriggerTypeDirect,
		Filters: map[string][]string{
			"provider": {"failing-ai"},
		},
		Providers: []*failover_strategy.ProviderConf{
			{Name: backupProvider, Model: backupModel},
		},
	})
	if err != nil {
		t.Fatalf("create failover handler error: %v", err)
	}

	ctx := newMockHttpContext()
	ctx.SetLabel("provider", "failing-ai")
	ctx.SetLabel("model", "gpt-3.5")

	chainExecuted := false
	chain := &mockChain{
		fn: func(c eocontext.EoContext) error {
			chainExecuted = true
			return nil
		},
	}

	err = strat.doFailover(ctx, chain, handler)
	if err != nil {
		t.Fatalf("doFailover error: %v", err)
	}

	if !convertReqCalled {
		t.Fatalf("expected RequestConvert to be called")
	}
	if !convertRespCalled {
		t.Fatalf("expected ResponseConvert to be called")
	}
	if !chainExecuted {
		t.Fatalf("expected next.DoChain to be called during failover")
	}

	// 验证注入的响应头
	if ctx.Response().GetHeader("Strategy-Failover") != "direct-failover-strat" {
		t.Fatalf("expected Strategy-Failover header to be direct-failover-strat, got %s",
			ctx.Response().GetHeader("Strategy-Failover"))
	}
	if ctx.Response().GetHeader("X-Failover-Provider") != backupProvider {
		t.Fatalf("expected X-Failover-Provider header to be %s, got %s",
			backupProvider, ctx.Response().GetHeader("X-Failover-Provider"))
	}
	if ctx.Response().GetHeader("X-AI-Provider") != backupProvider {
		t.Fatalf("expected X-AI-Provider header to be %s, got %s",
			backupProvider, ctx.Response().GetHeader("X-AI-Provider"))
	}
	if ctx.Response().GetHeader("X-AI-Model") != backupModel {
		t.Fatalf("expected X-AI-Model header to be %s, got %s",
			backupModel, ctx.Response().GetHeader("X-AI-Model"))
	}
}

func TestStrategy_ConditionFailover(t *testing.T) {
	strat := &Strategy{
		WorkerBase: drivers.Worker("failover-plugin-id", "failover-plugin"),
		modelType:  ai_convert.ModelTypeOpenAIChat,
	}

	backupProvider := "qwen"
	backupModel := "qwen-turbo"

	ai_convert.SetKeyResource(backupProvider, &mockKeyResource{
		id:       "qwen-key-1",
		priority: 1,
		health:   true,
		driver: &mockConverterDriver{
			provider:  backupProvider,
			modelType: ai_convert.ModelTypeOpenAIChat,
			requestConvertFn: func(ctx eocontext.EoContext, extender map[string]interface{}) error {
				return nil
			},
			responseConvertFn: func(ctx eocontext.EoContext) error {
				return nil
			},
		},
	})
	defer ai_convert.DelKeyResource(backupProvider, "qwen-key-1")

	handler, err := failover_strategy.NewFailoverHandler(&failover_strategy.Config{
		Name:        "condition-failover-strat",
		Priority:    1,
		TriggerType: failover_strategy.TriggerTypeCondition,
		Triggers: failover_strategy.TriggersConf{
			Failure: failover_strategy.TriggerFailureConf{
				Enabled:     true,
				StatusCodes: []int{500, 502, 503},
			},
		},
		Filters: map[string][]string{
			"provider": {"primary-provider"},
		},
		Providers: []*failover_strategy.ProviderConf{
			{Name: backupProvider, Model: backupModel},
		},
	})
	if err != nil {
		t.Fatalf("create handler error: %v", err)
	}

	ctx := newMockHttpContext()
	ctx.SetLabel("provider", "primary-provider")
	ctx.SetLabel("model", "primary-model")

	firstCall := true
	chain := &mockChain{
		fn: func(c eocontext.EoContext) error {
			if firstCall {
				firstCall = false
				ctx.Response().SetStatus(500, "Internal Server Error")
				return errors.New("primary upstream error")
			}
			// 灾备重试成功
			ctx.Response().SetStatus(200, "OK")
			return nil
		},
	}

	err = strat.doFailover(ctx, chain, handler)
	if err != nil {
		t.Fatalf("expected failover to succeed, got error: %v", err)
	}

	if ctx.Response().StatusCode() != 200 {
		t.Fatalf("expected status code 200, got %d", ctx.Response().StatusCode())
	}
	if ctx.Response().GetHeader("X-Failover-Provider") != backupProvider {
		t.Fatalf("expected X-Failover-Provider to be %s, got %s",
			backupProvider, ctx.Response().GetHeader("X-Failover-Provider"))
	}
}

func TestStrategy_BalancesFallback(t *testing.T) {
	strat := &Strategy{
		WorkerBase: drivers.Worker("failover-plugin-id", "failover-plugin"),
		modelType:  ai_convert.ModelTypeOpenAIChat,
	}

	balanceProvider := "deepseek"
	balanceModel := "deepseek-chat"

	// 注册全局 balance 供应商
	p := &mockProvider{
		id:       "balance-deepseek",
		provider: balanceProvider,
		model:    balanceModel,
		priority: 1,
		health:   true,
	}
	ai_convert.SetProvider(p)
	defer ai_convert.DelProvider("balance-deepseek")

	// 注册 key resource
	convertCalled := false
	ai_convert.SetKeyResource(balanceProvider, &mockKeyResource{
		id:       "deepseek-key-1",
		priority: 1,
		health:   true,
		driver: &mockConverterDriver{
			provider:  balanceProvider,
			modelType: ai_convert.ModelTypeOpenAIChat,
			requestConvertFn: func(ctx eocontext.EoContext, extender map[string]interface{}) error {
				convertCalled = true
				return nil
			},
			responseConvertFn: func(ctx eocontext.EoContext) error {
				return nil
			},
		},
	})
	defer ai_convert.DelKeyResource(balanceProvider, "deepseek-key-1")

	// 策略中配置的供应商无法使用（没有 key resource）
	handler, err := failover_strategy.NewFailoverHandler(&failover_strategy.Config{
		Name:        "balance-fallback-strat",
		Priority:    1,
		TriggerType: failover_strategy.TriggerTypeDirect,
		Filters: map[string][]string{
			"provider": {"broken-provider"},
		},
		Providers: []*failover_strategy.ProviderConf{
			{Name: "unconfigured-backup-provider", Model: "backup-model"},
		},
	})
	if err != nil {
		t.Fatalf("create handler error: %v", err)
	}

	ctx := newMockHttpContext()
	ctx.SetLabel("provider", "broken-provider")
	ctx.SetLabel("model", "chat-model")

	chain := &mockChain{
		fn: func(c eocontext.EoContext) error {
			ctx.Response().SetStatus(200, "OK")
			return nil
		},
	}

	err = strat.doFailover(ctx, chain, handler)
	if err != nil {
		t.Fatalf("expected balance fallback to succeed, got error: %v", err)
	}

	if !convertCalled {
		t.Fatalf("expected deepseek converter to be called")
	}
	if ctx.Response().GetHeader("X-AI-Provider") != balanceProvider {
		t.Fatalf("expected X-AI-Provider to be %s, got %s",
			balanceProvider, ctx.Response().GetHeader("X-AI-Provider"))
	}
}

func TestStrategy_FailoverExhausted_ReturnsBadRequest(t *testing.T) {
	strat := &Strategy{
		WorkerBase: drivers.Worker("failover-plugin-id", "failover-plugin"),
		modelType:  ai_convert.ModelTypeOpenAIChat,
	}

	handler, err := failover_strategy.NewFailoverHandler(&failover_strategy.Config{
		Name:        "exhausted-strat",
		Priority:    1,
		TriggerType: failover_strategy.TriggerTypeDirect,
		Filters: map[string][]string{
			"provider": {"exhausted-provider"},
		},
		Providers: []*failover_strategy.ProviderConf{
			{Name: "non-existent-provider", Model: "none"},
		},
	})
	if err != nil {
		t.Fatalf("create handler error: %v", err)
	}

	ctx := newMockHttpContext()
	ctx.SetLabel("provider", "exhausted-provider")
	ctx.SetLabel("model", "chat-model")

	chain := &mockChain{
		fn: func(c eocontext.EoContext) error {
			return nil
		},
	}

	err = strat.doFailover(ctx, chain, handler)
	if err == nil {
		t.Fatalf("expected error when all failovers are exhausted")
	}

	// 依照 ai-proxy executor.go#L217-219，当且仅当 status 不是 504 且 body 为空时，写入错误并置 400
	if ctx.Response().StatusCode() != 400 {
		t.Fatalf("expected status code 400, got %d", ctx.Response().StatusCode())
	}
	if len(ctx.Response().GetBody()) == 0 {
		t.Fatalf("expected non-empty response body")
	}
}

func TestStrategy_GatewayTimeout_Preserved(t *testing.T) {
	strat := &Strategy{
		WorkerBase: drivers.Worker("failover-plugin-id", "failover-plugin"),
		modelType:  ai_convert.ModelTypeOpenAIChat,
	}

	handler, err := failover_strategy.NewFailoverHandler(&failover_strategy.Config{
		Name:        "timeout-strat",
		Priority:    1,
		TriggerType: failover_strategy.TriggerTypeDirect,
		Filters: map[string][]string{
			"provider": {"timeout-provider"},
		},
		Providers: []*failover_strategy.ProviderConf{
			{Name: "non-existent-provider", Model: "none"},
		},
	})
	if err != nil {
		t.Fatalf("create handler error: %v", err)
	}

	ctx := newMockHttpContext()
	ctx.SetLabel("provider", "timeout-provider")
	ctx.SetLabel("model", "chat-model")
	ctx.Response().SetStatus(504, "Gateway Timeout")

	chain := &mockChain{
		fn: func(c eocontext.EoContext) error {
			return nil
		},
	}

	err = strat.doFailover(ctx, chain, handler)
	if err == nil {
		t.Fatalf("expected error")
	}

	// 依照 ai-proxy executor.go#L216，如果已经是 504，不能被篡改为 400
	if ctx.Response().StatusCode() != 504 {
		t.Fatalf("expected status code 504 to be preserved, got %d", ctx.Response().StatusCode())
	}
}

