package ai_proxy

import (
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	ai_convert "github.com/eolinker/apinto/ai-convert"
	context_label "github.com/eolinker/apinto/common/context-label"
	model_mapping "github.com/eolinker/apinto/drivers/model-mapping"
	failover_strategy "github.com/eolinker/apinto/drivers/strategy/failover-strategy"
	"github.com/eolinker/apinto/entries/ctx_key"
	http_context "github.com/eolinker/apinto/node/http-context"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/ohler55/ojg/jp"
	"github.com/valyala/fasthttp"
)

func newMockHttpContext(path string, body []byte) http_service.IHttpContext {
	fast := &fasthttp.RequestCtx{}
	freq := fasthttp.AcquireRequest()
	freq.SetRequestURI(path)
	if len(body) > 0 {
		freq.SetBody(body)
		freq.Header.SetContentType("application/json")
	}
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

type mockConverterDriver struct {
	provider          string
	modelType         ai_convert.ModelType
	requestConvertFn  func(ctx eocontext.EoContext, extender map[string]interface{}) error
	responseConvertFn func(ctx eocontext.EoContext) error
}

func (m *mockConverterDriver) Provider() string                { return m.provider }
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

type mockKeyResource struct {
	id     string
	driver ai_convert.IConverterDriver
}

func (m *mockKeyResource) ID() string      { return m.id }
func (m *mockKeyResource) Health() bool    { return true }
func (m *mockKeyResource) Priority() int   { return 1 }
func (m *mockKeyResource) Up()             {}
func (m *mockKeyResource) Down()           {}
func (m *mockKeyResource) IsBreaker() bool { return false }
func (m *mockKeyResource) Breaker()        {}
func (m *mockKeyResource) Get(modelType ai_convert.ModelType) (ai_convert.IConverterDriver, bool) {
	return m.driver, true
}
func (m *mockKeyResource) ModelTypeList() []ai_convert.ModelType {
	return []ai_convert.ModelType{ai_convert.ModelTypeOpenAIChat}
}

func TestAIProxy_DirectFailover(t *testing.T) {
	exec := &executor{
		modelType:   ai_convert.ModelTypeOpenAIChat,
		modelIdFrom: "path",
		config:      "{}",
	}

	backupProvider := "backup-ai"
	backupCalled := false

	ai_convert.SetKeyResource(backupProvider, &mockKeyResource{
		id: "backup-key",
		driver: &mockConverterDriver{
			provider:  backupProvider,
			modelType: ai_convert.ModelTypeOpenAIChat,
			requestConvertFn: func(ctx eocontext.EoContext, extender map[string]interface{}) error {
				backupCalled = true
				return nil
			},
		},
	})
	defer ai_convert.DelKeyResource(backupProvider, "backup-key")

	cfg := &failover_strategy.Config{
		Name:     "direct-failover-strat",
		Priority: 1,
		Direct: &failover_strategy.ProviderConf{
			Name: backupProvider,
		},
		Filters: map[string][]string{
			"provider": {"failing-ai"},
		},
	}
	handler, err := failover_strategy.NewHandler(cfg)
	if err != nil {
		t.Fatalf("create failover handler error: %v", err)
	}

	failover_strategy.SetStrategy(handler.Name(), handler, cfg.Filters)
	defer failover_strategy.DelStrategy(handler.Name())

	ctx := newMockHttpContext("/failing-ai/gpt-4", nil)
	chain := &mockChain{}

	err = exec.DoHttpFilter(ctx, chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !backupCalled {
		t.Fatalf("expected backup provider to be called")
	}

	if ctx.Response().GetHeader("Strategy-Failover") != "direct-failover-strat" {
		t.Fatalf("expected Strategy-Failover header, got %s", ctx.Response().GetHeader("Strategy-Failover"))
	}
	if ctx.Response().GetHeader("Strategy-Failover-Trigger-Type") != "direct" {
		t.Fatalf("expected Strategy-Failover-Trigger-Type header direct, got %s", ctx.Response().GetHeader("Strategy-Failover-Trigger-Type"))
	}
	if ctx.Response().GetHeader("Strategy-Failover-Trigger-Condition") != "direct" {
		t.Fatalf("expected Strategy-Failover-Trigger-Condition header direct, got %s", ctx.Response().GetHeader("Strategy-Failover-Trigger-Condition"))
	}
	if ctx.Response().GetHeader("Strategy-Failover-Provider") != backupProvider {
		t.Fatalf("expected Strategy-Failover-Provider header %s, got %s", backupProvider, ctx.Response().GetHeader("Strategy-Failover-Provider"))
	}
	if ctx.Response().GetHeader("Strategy-Failover-Original-Provider") != "failing-ai" {
		t.Fatalf("expected Strategy-Failover-Original-Provider header failing-ai, got %s", ctx.Response().GetHeader("Strategy-Failover-Original-Provider"))
	}
	if ctx.Response().GetHeader("X-AI-Provider") != backupProvider {
		t.Fatalf("expected X-AI-Provider header %s, got %s", backupProvider, ctx.Response().GetHeader("X-AI-Provider"))
	}
	if ctx.Response().GetHeader("X-AI-Model") != "gpt-4" {
		t.Fatalf("expected X-AI-Model header gpt-4, got %s", ctx.Response().GetHeader("X-AI-Model"))
	}
}

func TestAIProxy_ConditionFailover(t *testing.T) {
	exec := &executor{
		modelType:   ai_convert.ModelTypeOpenAIChat,
		modelIdFrom: "path",
		config:      "{}",
	}

	backupProvider := "backup-cond"
	backupCalled := false

	ai_convert.SetKeyResource(backupProvider, &mockKeyResource{
		id: "backup-cond-key",
		driver: &mockConverterDriver{
			provider:  backupProvider,
			modelType: ai_convert.ModelTypeOpenAIChat,
			requestConvertFn: func(ctx eocontext.EoContext, extender map[string]interface{}) error {
				backupCalled = true
				return nil
			},
		},
	})
	defer ai_convert.DelKeyResource(backupProvider, "backup-cond-key")

	cfg := &failover_strategy.Config{
		Name:     "cond-failover-strat",
		Priority: 1,
		Triggers: failover_strategy.TriggersConf{
			Failure: failover_strategy.TriggerFailureConf{
				Enabled: true,
			},
		},
		Filters: map[string][]string{
			"provider": {"primary-cond"},
		},
		Providers: []*failover_strategy.ProviderConf{
			{Name: backupProvider},
		},
	}
	handler, err := failover_strategy.NewHandler(cfg)
	if err != nil {
		t.Fatalf("create handler error: %v", err)
	}

	failover_strategy.SetStrategy(handler.Name(), handler, cfg.Filters)
	defer failover_strategy.DelStrategy(handler.Name())

	firstCall := true
	chain := &mockChain{
		fn: func(c eocontext.EoContext) error {
			if firstCall {
				firstCall = false
				ctx := c.(http_service.IHttpContext)
				ctx.Response().SetStatus(500, "Internal Server Error")
				return errors.New("upstream failure")
			}
			return nil
		},
	}

	ctx := newMockHttpContext("/primary-cond/chat", nil)
	err = exec.DoHttpFilter(ctx, chain)
	if err != nil {
		t.Fatalf("expected failover to succeed, got error: %v", err)
	}

	if !backupCalled {
		t.Fatalf("expected backup provider to be called on condition failure")
	}
	if ctx.Response().GetHeader("Strategy-Failover") != "cond-failover-strat" {
		t.Fatalf("expected Strategy-Failover header, got %s", ctx.Response().GetHeader("Strategy-Failover"))
	}
	if ctx.Response().GetHeader("Strategy-Failover-Trigger-Type") != "conditional" {
		t.Fatalf("expected Strategy-Failover-Trigger-Type header conditional, got %s", ctx.Response().GetHeader("Strategy-Failover-Trigger-Type"))
	}
	if ctx.Response().GetHeader("Strategy-Failover-Trigger-Condition") != "failure" {
		t.Fatalf("expected Strategy-Failover-Trigger-Condition header failure, got %s", ctx.Response().GetHeader("Strategy-Failover-Trigger-Condition"))
	}
	if ctx.Response().GetHeader("Strategy-Failover-Provider") != backupProvider {
		t.Fatalf("expected Strategy-Failover-Provider header %s, got %s", backupProvider, ctx.Response().GetHeader("Strategy-Failover-Provider"))
	}
	if ctx.Response().GetHeader("Strategy-Failover-Original-Provider") != "primary-cond" {
		t.Fatalf("expected Strategy-Failover-Original-Provider header primary-cond, got %s", ctx.Response().GetHeader("Strategy-Failover-Original-Provider"))
	}
	if ctx.Response().GetHeader("X-AI-Provider") != backupProvider {
		t.Fatalf("expected X-AI-Provider header %s, got %s", backupProvider, ctx.Response().GetHeader("X-AI-Provider"))
	}
}

func TestAIProxy_TimeoutInterruption(t *testing.T) {
	exec := &executor{
		modelType:   ai_convert.ModelTypeOpenAIChat,
		modelIdFrom: "path",
		config:      "{}",
	}

	backupProvider := "backup-timeout"
	backupCalled := false

	ai_convert.SetKeyResource(backupProvider, &mockKeyResource{
		id: "backup-timeout-key",
		driver: &mockConverterDriver{
			provider:  backupProvider,
			modelType: ai_convert.ModelTypeOpenAIChat,
			requestConvertFn: func(ctx eocontext.EoContext, extender map[string]interface{}) error {
				backupCalled = true
				return nil
			},
		},
	})
	defer ai_convert.DelKeyResource(backupProvider, "backup-timeout-key")

	cfg := &failover_strategy.Config{
		Name:     "timeout-failover-strat",
		Priority: 1,
		Triggers: failover_strategy.TriggersConf{
			Timeout: failover_strategy.TriggerTimeoutConf{
				Enabled:        true,
				TimeoutSeconds: 1,
			},
		},
		Filters: map[string][]string{
			"provider": {"timeout-primary"},
		},
		Providers: []*failover_strategy.ProviderConf{
			{Name: backupProvider},
		},
	}
	handler, err := failover_strategy.NewHandler(cfg)
	if err != nil {
		t.Fatalf("create handler error: %v", err)
	}

	failover_strategy.SetStrategy(handler.Name(), handler, cfg.Filters)
	defer failover_strategy.DelStrategy(handler.Name())

	firstCall := true
	chain := &mockChain{
		fn: func(c eocontext.EoContext) error {
			if firstCall {
				firstCall = false
				time.Sleep(1200 * time.Millisecond)
				return nil
			}
			if httpCtx, ok := c.(http_service.IHttpContext); ok {
				httpCtx.Response().SetStatus(200, "OK")
			}
			return nil
		},
	}

	ctx := newMockHttpContext("/timeout-primary/model", nil)
	err = exec.DoHttpFilter(ctx, chain)
	if err != nil {
		t.Fatalf("expected failover to succeed after timeout, got: %v", err)
	}

	if !backupCalled {
		t.Fatalf("expected backup provider to be called after timeout")
	}
	if ctx.Response().GetHeader("Strategy-Failover") != "timeout-failover-strat" {
		t.Fatalf("expected Strategy-Failover header, got %s", ctx.Response().GetHeader("Strategy-Failover"))
	}
}

func TestAIProxy_TimeoutInterruption_TriggerDisabledNoFallback(t *testing.T) {
	exec := &executor{
		modelType:   ai_convert.ModelTypeOpenAIChat,
		modelIdFrom: "path",
		config:      "{}",
	}

	backupProvider := "disabled-backup"
	backupCalled := false

	ai_convert.SetKeyResource(backupProvider, &mockKeyResource{
		id: "disabled-backup-key",
		driver: &mockConverterDriver{
			provider:  backupProvider,
			modelType: ai_convert.ModelTypeOpenAIChat,
			requestConvertFn: func(ctx eocontext.EoContext, extender map[string]interface{}) error {
				backupCalled = true
				return nil
			},
		},
	})
	defer ai_convert.DelKeyResource(backupProvider, "disabled-backup-key")

	cfg := &failover_strategy.Config{
		Name:     "disabled-timeout-strat",
		Priority: 1,
		Triggers: failover_strategy.TriggersConf{
			Timeout: failover_strategy.TriggerTimeoutConf{
				Enabled:        false,
				TimeoutSeconds: 1,
			},
			Failure: failover_strategy.TriggerFailureConf{
				Enabled: false,
			},
		},
		Filters: map[string][]string{
			"provider": {"slow-provider"},
		},
		Providers: []*failover_strategy.ProviderConf{
			{Name: backupProvider},
		},
	}
	handler, err := failover_strategy.NewHandler(cfg)
	if err != nil {
		t.Fatalf("create handler error: %v", err)
	}

	failover_strategy.SetStrategy(handler.Name(), handler, cfg.Filters)
	defer failover_strategy.DelStrategy(handler.Name())

	chain := &mockChain{
		fn: func(c eocontext.EoContext) error {
			time.Sleep(1200 * time.Millisecond)
			return nil
		},
	}

	ctx := newMockHttpContext("/slow-provider/model", nil)
	err = exec.DoHttpFilter(ctx, chain)
	if err == nil {
		t.Fatalf("expected error due to timeout interruption")
	}

	if backupCalled {
		t.Fatalf("backup provider should NOT be called when timeout trigger is disabled")
	}

	if ctx.Response().StatusCode() != http.StatusGatewayTimeout {
		t.Fatalf("expected status code 504 Gateway Timeout, got %d", ctx.Response().StatusCode())
	}

	if !context_label.IsAITimeout(ctx) {
		t.Fatalf("expected ai_timeout context label to be set")
	}
}

func TestAIProxy_ResourceMatchPriority(t *testing.T) {
	exec := &executor{
		modelType:   ai_convert.ModelTypeOpenAIChat,
		modelIdFrom: "path",
		config:      "{}",
	}

	providerBackup := "provider-backup"
	resourceBackup := "resource-backup"

	var lastCalledProvider string

	ai_convert.SetKeyResource(providerBackup, &mockKeyResource{
		id: "provider-backup-key",
		driver: &mockConverterDriver{
			provider:  providerBackup,
			modelType: ai_convert.ModelTypeOpenAIChat,
			requestConvertFn: func(ctx eocontext.EoContext, extender map[string]interface{}) error {
				lastCalledProvider = providerBackup
				return nil
			},
		},
	})
	defer ai_convert.DelKeyResource(providerBackup, "provider-backup-key")

	ai_convert.SetKeyResource(resourceBackup, &mockKeyResource{
		id: "resource-backup-key",
		driver: &mockConverterDriver{
			provider:  resourceBackup,
			modelType: ai_convert.ModelTypeOpenAIChat,
			requestConvertFn: func(ctx eocontext.EoContext, extender map[string]interface{}) error {
				lastCalledProvider = resourceBackup
				return nil
			},
		},
	})
	defer ai_convert.DelKeyResource(resourceBackup, "resource-backup-key")

	// 策略 1: provider 级别匹配
	cfgProvider := &failover_strategy.Config{
		Name:     "provider-strat",
		Priority: 10,
		Direct: &failover_strategy.ProviderConf{
			Name: providerBackup,
		},
		Filters: map[string][]string{
			"provider": {"ai-svc"},
		},
	}
	hProvider, err := failover_strategy.NewHandler(cfgProvider)
	if err != nil {
		t.Fatalf("create handler error: %v", err)
	}
	failover_strategy.SetStrategy(hProvider.Name(), hProvider, cfgProvider.Filters)
	defer failover_strategy.DelStrategy(hProvider.Name())

	// 策略 2: resource 级别精准匹配
	cfgResource := &failover_strategy.Config{
		Name:     "resource-strat",
		Priority: 1,
		Direct: &failover_strategy.ProviderConf{
			Name: resourceBackup,
		},
		Filters: map[string][]string{
			"resource": {"ai-svc/special-model"},
		},
	}
	hResource, err := failover_strategy.NewHandler(cfgResource)
	if err != nil {
		t.Fatalf("create handler error: %v", err)
	}
	failover_strategy.SetStrategy(hResource.Name(), hResource, cfgResource.Filters)
	defer failover_strategy.DelStrategy(hResource.Name())

	chain := &mockChain{}

	// 请求 1：访问精确匹配的 resource，应当优先命中 resource-strat（即使 priority 较低）
	ctx1 := newMockHttpContext("/ai-svc/special-model", nil)
	err = exec.DoHttpFilter(ctx1, chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lastCalledProvider != resourceBackup {
		t.Fatalf("expected resource-backup to be called for resource match, got %s", lastCalledProvider)
	}

	// 请求 2：访问非 special-model，应当回退命中 provider-strat
	ctx2 := newMockHttpContext("/ai-svc/other-model", nil)
	err = exec.DoHttpFilter(ctx2, chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lastCalledProvider != providerBackup {
		t.Fatalf("expected provider-backup to be called for general model, got %s", lastCalledProvider)
	}
}

func TestAIProxy_DirectAndConditionFailover(t *testing.T) {
	exec := &executor{
		modelType:   ai_convert.ModelTypeOpenAIChat,
		modelIdFrom: "path",
		config:      "{}",
	}

	directProvider := "direct-ai"
	fallbackProvider := "fallback-ai"
	directCalled := false
	fallbackCalled := false

	ai_convert.SetKeyResource(directProvider, &mockKeyResource{
		id: "direct-ai-key",
		driver: &mockConverterDriver{
			provider:  directProvider,
			modelType: ai_convert.ModelTypeOpenAIChat,
			requestConvertFn: func(ctx eocontext.EoContext, extender map[string]interface{}) error {
				directCalled = true
				return nil
			},
		},
	})
	defer ai_convert.DelKeyResource(directProvider, "direct-ai-key")

	ai_convert.SetKeyResource(fallbackProvider, &mockKeyResource{
		id: "fallback-ai-key",
		driver: &mockConverterDriver{
			provider:  fallbackProvider,
			modelType: ai_convert.ModelTypeOpenAIChat,
			requestConvertFn: func(ctx eocontext.EoContext, extender map[string]interface{}) error {
				fallbackCalled = true
				return nil
			},
		},
	})
	defer ai_convert.DelKeyResource(fallbackProvider, "fallback-ai-key")

	cfg := &failover_strategy.Config{
		Name:     "direct-and-cond-strat",
		Priority: 1,
		Direct: &failover_strategy.ProviderConf{
			Name: directProvider,
		},
		Triggers: failover_strategy.TriggersConf{
			Failure: failover_strategy.TriggerFailureConf{
				Enabled: true,
			},
		},
		Filters: map[string][]string{
			"provider": {"origin-ai"},
		},
		Providers: []*failover_strategy.ProviderConf{
			{Name: fallbackProvider},
		},
	}
	handler, err := failover_strategy.NewHandler(cfg)
	if err != nil {
		t.Fatalf("create handler error: %v", err)
	}

	failover_strategy.SetStrategy(handler.Name(), handler, cfg.Filters)
	defer failover_strategy.DelStrategy(handler.Name())

	firstCall := true
	chain := &mockChain{
		fn: func(c eocontext.EoContext) error {
			if firstCall {
				firstCall = false
				ctx := c.(http_service.IHttpContext)
				ctx.Response().SetStatus(500, "Internal Server Error")
				return errors.New("direct upstream failure")
			}
			return nil
		},
	}

	ctx := newMockHttpContext("/origin-ai/chat-model", nil)
	err = exec.DoHttpFilter(ctx, chain)
	if err != nil {
		t.Fatalf("expected failover to succeed, got error: %v", err)
	}

	if !directCalled {
		t.Fatalf("expected direct provider to be called first")
	}
	if !fallbackCalled {
		t.Fatalf("expected fallback provider to be called after direct provider failed")
	}

	if ctx.Response().GetHeader("Strategy-Failover") != "direct-and-cond-strat" {
		t.Fatalf("expected Strategy-Failover header, got %s", ctx.Response().GetHeader("Strategy-Failover"))
	}
	if ctx.Response().GetHeader("Strategy-Failover-Direct") != directProvider {
		t.Fatalf("expected Strategy-Failover-Direct header %s, got %s", directProvider, ctx.Response().GetHeader("Strategy-Failover-Direct"))
	}
	if ctx.Response().GetHeader("Strategy-Failover-Trigger-Type") != "conditional" {
		t.Fatalf("expected Strategy-Failover-Trigger-Type header conditional, got %s", ctx.Response().GetHeader("Strategy-Failover-Trigger-Type"))
	}
	if ctx.Response().GetHeader("Strategy-Failover-Trigger-Condition") != "failure" {
		t.Fatalf("expected Strategy-Failover-Trigger-Condition header failure, got %s", ctx.Response().GetHeader("Strategy-Failover-Trigger-Condition"))
	}
	if ctx.Response().GetHeader("Strategy-Failover-Original-Provider") != "origin-ai" {
		t.Fatalf("expected Strategy-Failover-Original-Provider header origin-ai, got %s", ctx.Response().GetHeader("Strategy-Failover-Original-Provider"))
	}
	if ctx.Response().GetHeader("Strategy-Failover-Provider") != fallbackProvider {
		t.Fatalf("expected Strategy-Failover-Provider header %s, got %s", fallbackProvider, ctx.Response().GetHeader("Strategy-Failover-Provider"))
	}
	if ctx.Response().GetHeader("X-AI-Provider") != fallbackProvider {
		t.Fatalf("expected X-AI-Provider header %s, got %s", fallbackProvider, ctx.Response().GetHeader("X-AI-Provider"))
	}
}

func TestAIProxy_FallbackProviderFollowsTriggerCondition(t *testing.T) {
	exec := &executor{
		modelType:   ai_convert.ModelTypeOpenAIChat,
		modelIdFrom: "path",
		config:      "{}",
	}

	fb1Provider := "fallback-ai-1"
	fb2Provider := "fallback-ai-2"
	fb1Called := false
	fb2Called := false

	ai_convert.SetKeyResource(fb1Provider, &mockKeyResource{
		id: "fb1-key",
		driver: &mockConverterDriver{
			provider:  fb1Provider,
			modelType: ai_convert.ModelTypeOpenAIChat,
			requestConvertFn: func(ctx eocontext.EoContext, extender map[string]interface{}) error {
				fb1Called = true
				return nil
			},
		},
	})
	defer ai_convert.DelKeyResource(fb1Provider, "fb1-key")

	ai_convert.SetKeyResource(fb2Provider, &mockKeyResource{
		id: "fb2-key",
		driver: &mockConverterDriver{
			provider:  fb2Provider,
			modelType: ai_convert.ModelTypeOpenAIChat,
			requestConvertFn: func(ctx eocontext.EoContext, extender map[string]interface{}) error {
				fb2Called = true
				return nil
			},
		},
	})
	defer ai_convert.DelKeyResource(fb2Provider, "fb2-key")

	cfg := &failover_strategy.Config{
		Name:     "fb-chain-strat",
		Priority: 1,
		Triggers: failover_strategy.TriggersConf{
			Failure: failover_strategy.TriggerFailureConf{
				Enabled: true,
			},
		},
		Filters: map[string][]string{
			"provider": {"origin-ai-fb"},
		},
		Providers: []*failover_strategy.ProviderConf{
			{Name: fb1Provider},
			{Name: fb2Provider},
		},
	}
	handler, err := failover_strategy.NewHandler(cfg)
	if err != nil {
		t.Fatalf("create handler error: %v", err)
	}

	failover_strategy.SetStrategy(handler.Name(), handler, cfg.Filters)
	defer failover_strategy.DelStrategy(handler.Name())

	callCount := 0
	chain := &mockChain{
		fn: func(c eocontext.EoContext) error {
			callCount++
			ctx := c.(http_service.IHttpContext)
			if callCount == 1 {
				// 原始上游报错
				ctx.Response().SetStatus(500, "Internal Server Error")
				return errors.New("primary error")
			}
			if callCount == 2 {
				// 灾备供应商 1 同样报错（符合触发条件）
				ctx.Response().SetStatus(502, "Bad Gateway")
				return errors.New("fallback 1 upstream error")
			}
			// 灾备供应商 2 成功
			ctx.Response().SetStatus(200, "OK")
			return nil
		},
	}

	ctx := newMockHttpContext("/origin-ai-fb/chat-model", nil)
	err = exec.DoHttpFilter(ctx, chain)
	if err != nil {
		t.Fatalf("expected fallback to succeed at fb2, got error: %v", err)
	}

	if !fb1Called {
		t.Fatalf("expected fallback-ai-1 to be called")
	}
	if !fb2Called {
		t.Fatalf("expected fallback-ai-2 to be called after fallback-ai-1 triggered failover condition")
	}
	if callCount != 3 {
		t.Fatalf("expected 3 total calls, got %d", callCount)
	}

	if ctx.Response().GetHeader("Strategy-Failover-Provider") != fb2Provider {
		t.Fatalf("expected Strategy-Failover-Provider header %s, got %s", fb2Provider, ctx.Response().GetHeader("Strategy-Failover-Provider"))
	}
	if ctx.Response().GetHeader("X-AI-Provider") != fb2Provider {
		t.Fatalf("expected X-AI-Provider header %s, got %s", fb2Provider, ctx.Response().GetHeader("X-AI-Provider"))
	}
}

func TestAIProxy_AllFallbackProvidersTriggerFailure(t *testing.T) {
	exec := &executor{
		modelType:   ai_convert.ModelTypeOpenAIChat,
		modelIdFrom: "path",
		config:      "{}",
	}

	fb1Provider := "all-fail-1"
	fb2Provider := "all-fail-2"
	fb1Called := false
	fb2Called := false

	ai_convert.SetKeyResource(fb1Provider, &mockKeyResource{
		id: "all-fail-1-key",
		driver: &mockConverterDriver{
			provider:  fb1Provider,
			modelType: ai_convert.ModelTypeOpenAIChat,
			requestConvertFn: func(ctx eocontext.EoContext, extender map[string]interface{}) error {
				fb1Called = true
				return nil
			},
		},
	})
	defer ai_convert.DelKeyResource(fb1Provider, "all-fail-1-key")

	ai_convert.SetKeyResource(fb2Provider, &mockKeyResource{
		id: "all-fail-2-key",
		driver: &mockConverterDriver{
			provider:  fb2Provider,
			modelType: ai_convert.ModelTypeOpenAIChat,
			requestConvertFn: func(ctx eocontext.EoContext, extender map[string]interface{}) error {
				fb2Called = true
				return nil
			},
		},
	})
	defer ai_convert.DelKeyResource(fb2Provider, "all-fail-2-key")

	cfg := &failover_strategy.Config{
		Name:     "all-fail-strat",
		Priority: 1,
		Triggers: failover_strategy.TriggersConf{
			Failure: failover_strategy.TriggerFailureConf{
				Enabled: true,
			},
		},
		Filters: map[string][]string{
			"provider": {"all-fail-origin"},
		},
		Providers: []*failover_strategy.ProviderConf{
			{Name: fb1Provider},
			{Name: fb2Provider},
		},
	}
	handler, err := failover_strategy.NewHandler(cfg)
	if err != nil {
		t.Fatalf("create handler error: %v", err)
	}

	failover_strategy.SetStrategy(handler.Name(), handler, cfg.Filters)
	defer failover_strategy.DelStrategy(handler.Name())

	chain := &mockChain{
		fn: func(c eocontext.EoContext) error {
			ctx := c.(http_service.IHttpContext)
			ctx.Response().SetStatus(500, "Internal Server Error")
			return errors.New("continuous failure")
		},
	}

	ctx := newMockHttpContext("/all-fail-origin/chat", nil)
	err = exec.DoHttpFilter(ctx, chain)
	if err == nil {
		t.Fatalf("expected error when all providers fail, got nil")
	}

	if !fb1Called || !fb2Called {
		t.Fatalf("expected all fallback providers to have been attempted (fb1: %v, fb2: %v)", fb1Called, fb2Called)
	}
}

func TestAIProxy_UpstreamTimingBaseline(t *testing.T) {
	exec := &executor{
		modelType:   ai_convert.ModelTypeOpenAIChat,
		modelIdFrom: "path",
		config:      "{}",
	}

	backupProvider := "upstream-timed-backup"
	backupCalled := false

	ai_convert.SetKeyResource(backupProvider, &mockKeyResource{
		id: "upstream-backup-key",
		driver: &mockConverterDriver{
			provider:  backupProvider,
			modelType: ai_convert.ModelTypeOpenAIChat,
			requestConvertFn: func(ctx eocontext.EoContext, extender map[string]interface{}) error {
				backupCalled = true
				return nil
			},
		},
	})
	defer ai_convert.DelKeyResource(backupProvider, "upstream-backup-key")

	cfg := &failover_strategy.Config{
		Name:     "upstream-time-strat",
		Priority: 1,
		Triggers: failover_strategy.TriggersConf{
			Timeout: failover_strategy.TriggerTimeoutConf{
				Enabled:        true,
				TimeoutSeconds: 1, // 1秒超时
			},
		},
		Filters: map[string][]string{
			"provider": {"upstream-time-primary"},
		},
		Providers: []*failover_strategy.ProviderConf{
			{Name: backupProvider},
		},
	}
	handler, err := failover_strategy.NewHandler(cfg)
	if err != nil {
		t.Fatalf("create handler error: %v", err)
	}

	failover_strategy.SetStrategy(handler.Name(), handler, cfg.Filters)
	defer failover_strategy.DelStrategy(handler.Name())

	// 测试：上游真实耗时仅 100ms，未超 1s，不触发超时灾备
	chainNoTimeout := &mockChain{
		fn: func(c eocontext.EoContext) error {
			ctx := c.(http_service.IHttpContext)
			// 设置真实上游网络耗时为 100ms
			context_label.SetUpstreamCost(ctx, 100*time.Millisecond)
			ctx.Response().SetStatus(200, "OK")
			return nil
		},
	}

	ctxNoTimeout := newMockHttpContext("/upstream-time-primary/model", nil)
	err = exec.DoHttpFilter(ctxNoTimeout, chainNoTimeout)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if backupCalled {
		t.Fatalf("upstream cost 100ms should NOT trigger timeout failover")
	}

	// 测试：上游真实耗时 1200ms，超过 1s，触发超时灾备
	chainTimeout := &mockChain{
		fn: func(c eocontext.EoContext) error {
			ctx := c.(http_service.IHttpContext)
			// 第一次调用设置上游网络耗时 1200ms
			if !backupCalled {
				context_label.SetUpstreamCost(ctx, 1200*time.Millisecond)
				time.Sleep(1100 * time.Millisecond)
				ctx.Response().SetStatus(200, "OK")
				return nil
			}
			context_label.SetUpstreamCost(ctx, 50*time.Millisecond)
			ctx.Response().SetStatus(200, "OK")
			return nil
		},
	}

	ctxTimeout := newMockHttpContext("/upstream-time-primary/model", nil)
	err = exec.DoHttpFilter(ctxTimeout, chainTimeout)
	if err != nil {
		t.Fatalf("expected failover to succeed, got error: %v", err)
	}
	if !backupCalled {
		t.Fatalf("upstream cost 1200ms SHOULD trigger timeout failover")
	}
	if ctxTimeout.Response().GetHeader("Strategy-Failover") != "upstream-time-strat" {
		t.Fatalf("expected Strategy-Failover header, got %s", ctxTimeout.Response().GetHeader("Strategy-Failover"))
	}
}

func TestAIProxy_TimeoutAutoInterruptionAndRetry(t *testing.T) {
	exec := &executor{
		modelType:   ai_convert.ModelTypeOpenAIChat,
		modelIdFrom: "path",
		config:      "{}",
	}

	backupProvider := "auto-interrupt-backup"
	backupCalled := false

	ai_convert.SetKeyResource(backupProvider, &mockKeyResource{
		id: "auto-interrupt-backup-key",
		driver: &mockConverterDriver{
			provider:  backupProvider,
			modelType: ai_convert.ModelTypeOpenAIChat,
			requestConvertFn: func(ctx eocontext.EoContext, extender map[string]interface{}) error {
				backupCalled = true
				return nil
			},
		},
	})
	defer ai_convert.DelKeyResource(backupProvider, "auto-interrupt-backup-key")

	cfg := &failover_strategy.Config{
		Name:     "auto-interrupt-strat",
		Priority: 1,
		Triggers: failover_strategy.TriggersConf{
			Timeout: failover_strategy.TriggerTimeoutConf{
				Enabled:        true,
				TimeoutSeconds: 1, // 1秒策略超时
			},
		},
		Filters: map[string][]string{
			"provider": {"hanging-primary"},
		},
		Providers: []*failover_strategy.ProviderConf{
			{Name: backupProvider},
		},
	}
	handler, err := failover_strategy.NewHandler(cfg)
	if err != nil {
		t.Fatalf("create handler error: %v", err)
	}

	failover_strategy.SetStrategy(handler.Name(), handler, cfg.Filters)
	defer failover_strategy.DelStrategy(handler.Name())

	firstCall := true
	chain := &mockChain{
		fn: func(c eocontext.EoContext) error {
			if firstCall {
				firstCall = false
				// 验证策略通过 ctx_key.CtxKeyTimeout 设置并修改了超时时间
				tVal := c.Value(ctx_key.CtxKeyTimeout)
				timeout, ok := tVal.(time.Duration)
				if !ok || timeout != time.Second {
					t.Errorf("expected ctx_key.CtxKeyTimeout to be 1s, got %v", tVal)
				}
				// 模拟请求超过策略设置的超时时间，底层自动中断并返回 fasthttp.ErrTimeout
				httpCtx := c.(http_service.IHttpContext)
				context_label.SetAITimeout(httpCtx, true)
				httpCtx.Response().SetStatus(504, "Gateway Timeout")
				return fasthttp.ErrTimeout
			}
			httpCtx := c.(http_service.IHttpContext)
			httpCtx.Response().SetStatus(200, "OK")
			return nil
		},
	}

	ctx := newMockHttpContext("/hanging-primary/model", nil)
	err = exec.DoHttpFilter(ctx, chain)

	if err != nil {
		t.Fatalf("expected failover retry to succeed after auto interruption, got error: %v", err)
	}
	if !backupCalled {
		t.Fatalf("expected backup provider to be called during retry")
	}

	if ctx.Response().GetHeader("Strategy-Failover") != "auto-interrupt-strat" {
		t.Fatalf("expected Strategy-Failover header auto-interrupt-strat, got %s", ctx.Response().GetHeader("Strategy-Failover"))
	}
	if ctx.Response().GetHeader("Strategy-Failover-Provider") != backupProvider {
		t.Fatalf("expected Strategy-Failover-Provider header %s, got %s", backupProvider, ctx.Response().GetHeader("Strategy-Failover-Provider"))
	}
}

func TestAIProxy_ChineseHeaderEncoding(t *testing.T) {
	chineseProvider := "备选供应商-1"
	escaped := encodeHeaderValue(chineseProvider)
	if !strings.Contains(escaped, "%") {
		t.Fatalf("expected escaped string to contain percent-encoding, got %s", escaped)
	}
	unescaped, err := url.QueryUnescape(escaped)
	if err != nil {
		t.Fatalf("query unescape error: %v", err)
	}
	if unescaped != chineseProvider {
		t.Fatalf("expected unescaped to match %s, got %s", chineseProvider, unescaped)
	}

	// 纯 ASCII 字符保持原样
	asciiProvider := "openai-provider-1"
	if encodeHeaderValue(asciiProvider) != asciiProvider {
		t.Fatalf("expected ascii to stay unchanged, got %s", encodeHeaderValue(asciiProvider))
	}
}

func TestAIProxy_FailureStatusCode401And403(t *testing.T) {
	backupProvider1 := "status-backup-1"
	backupProvider2 := "status-backup-2"

	var backup1Called, backup2Called bool

	exec := &executor{
		modelType:   ai_convert.ModelTypeOpenAIChat,
		modelIdFrom: "path",
		config:      "{}",
	}

	ai_convert.SetKeyResource(backupProvider1, &mockKeyResource{
		id: "status-backup-key-1",
		driver: &mockConverterDriver{
			provider:  backupProvider1,
			modelType: ai_convert.ModelTypeOpenAIChat,
			requestConvertFn: func(ctx eocontext.EoContext, extender map[string]interface{}) error {
				backup1Called = true
				return nil
			},
		},
	})
	defer ai_convert.DelKeyResource(backupProvider1, "status-backup-key-1")

	ai_convert.SetKeyResource(backupProvider2, &mockKeyResource{
		id: "status-backup-key-2",
		driver: &mockConverterDriver{
			provider:  backupProvider2,
			modelType: ai_convert.ModelTypeOpenAIChat,
			requestConvertFn: func(ctx eocontext.EoContext, extender map[string]interface{}) error {
				backup2Called = true
				return nil
			},
		},
	})
	defer ai_convert.DelKeyResource(backupProvider2, "status-backup-key-2")

	cfg := &failover_strategy.Config{
		Name:     "status-failover-strat",
		Priority: 10,
		Triggers: failover_strategy.TriggersConf{
			Failure: failover_strategy.TriggerFailureConf{
				Enabled: true,
			},
		},
		Filters: map[string][]string{
			"provider": {"status-primary"},
		},
		Providers: []*failover_strategy.ProviderConf{
			{Name: backupProvider1},
			{Name: backupProvider2},
		},
	}
	handler, err := failover_strategy.NewHandler(cfg)
	if err != nil {
		t.Fatalf("create handler error: %v", err)
	}

	failover_strategy.SetStrategy(handler.Name(), handler, cfg.Filters)
	defer failover_strategy.DelStrategy(handler.Name())

	// 测试：Primary 返回 401 Unauthorized，Backup 1 返回 403 Forbidden，应继续灾备重试到 Backup 2 成功 (200)
	chain := &mockChain{
		fn: func(c eocontext.EoContext) error {
			httpCtx := c.(http_service.IHttpContext)
			currProvider := ai_convert.GetAIProvider(httpCtx)
			if currProvider == "status-primary" {
				httpCtx.Response().SetStatus(http.StatusUnauthorized, "Unauthorized")
				httpCtx.Response().SetBody([]byte(`{"error":"invalid_api_key"}`))
				return nil
			}
			if currProvider == backupProvider1 {
				httpCtx.Response().SetStatus(http.StatusForbidden, "Forbidden")
				httpCtx.Response().SetBody([]byte(`{"error":"access_denied"}`))
				return nil
			}
			if currProvider == backupProvider2 {
				httpCtx.Response().SetStatus(http.StatusOK, "OK")
				httpCtx.Response().SetBody([]byte(`{"result":"ok"}`))
				return nil
			}
			return nil
		},
	}

	ctx := newMockHttpContext("/status-primary/model", nil)
	err = exec.DoHttpFilter(ctx, chain)
	if err != nil {
		t.Fatalf("expected failover to succeed, got error: %v", err)
	}
	if !backup1Called {
		t.Fatalf("expected backup 1 to be called")
	}
	if !backup2Called {
		t.Fatalf("expected backup 2 to be called after backup 1 returned 403")
	}
	if ctx.Response().StatusCode() != http.StatusOK {
		t.Fatalf("expected final status 200, got %d", ctx.Response().StatusCode())
	}
	if ctx.Response().GetHeader("Strategy-Failover-Provider") != backupProvider2 {
		t.Fatalf("expected Strategy-Failover-Provider to be %s, got %s", backupProvider2, ctx.Response().GetHeader("Strategy-Failover-Provider"))
	}
	if ctx.Response().GetHeader("Strategy-Failover-Trigger-Type") != "conditional" {
		t.Fatalf("expected Strategy-Failover-Trigger-Type to be conditional, got %s", ctx.Response().GetHeader("Strategy-Failover-Trigger-Type"))
	}
	if ctx.Response().GetHeader("Strategy-Failover-Trigger-Condition") != "failure" {
		t.Fatalf("expected Strategy-Failover-Trigger-Condition to be failure, got %s", ctx.Response().GetHeader("Strategy-Failover-Trigger-Condition"))
	}
	if ctx.Response().GetHeader("Strategy-Failover-Original-Provider") != "status-primary" {
		t.Fatalf("expected Strategy-Failover-Original-Provider to be status-primary, got %s", ctx.Response().GetHeader("Strategy-Failover-Original-Provider"))
	}
}

func TestAIProxy_FailoverNoTriggerNoHeaders(t *testing.T) {
	exec := &executor{
		modelType:   ai_convert.ModelTypeOpenAIChat,
		modelIdFrom: "path",
		config:      "{}",
	}

	backupProvider := "unused-backup"
	ai_convert.SetKeyResource(backupProvider, &mockKeyResource{
		id: "unused-backup-key",
		driver: &mockConverterDriver{
			provider:  backupProvider,
			modelType: ai_convert.ModelTypeOpenAIChat,
		},
	})
	defer ai_convert.DelKeyResource(backupProvider, "unused-backup-key")

	cfg := &failover_strategy.Config{
		Name:     "normal-strat",
		Priority: 1,
		Triggers: failover_strategy.TriggersConf{
			Failure: failover_strategy.TriggerFailureConf{
				Enabled: true,
			},
		},
		Filters: map[string][]string{
			"provider": {"normal-primary"},
		},
		Providers: []*failover_strategy.ProviderConf{
			{Name: backupProvider},
		},
	}
	handler, err := failover_strategy.NewHandler(cfg)
	if err != nil {
		t.Fatalf("create handler error: %v", err)
	}

	failover_strategy.SetStrategy(handler.Name(), handler, cfg.Filters)
	defer failover_strategy.DelStrategy(handler.Name())

	chain := &mockChain{
		fn: func(c eocontext.EoContext) error {
			httpCtx := c.(http_service.IHttpContext)
			httpCtx.Response().SetStatus(http.StatusOK, "OK")
			return nil
		},
	}

	ctx := newMockHttpContext("/normal-primary/model-1", nil)
	err = exec.DoHttpFilter(ctx, chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 正常未触发灾备，不应存在灾备触发类型和灾备生效供应商等头
	if triggerType := ctx.Response().GetHeader("Strategy-Failover-Trigger-Type"); triggerType != "" {
		t.Fatalf("expected no Strategy-Failover-Trigger-Type when failover not triggered, got %s", triggerType)
	}
	if failoverProv := ctx.Response().GetHeader("Strategy-Failover-Provider"); failoverProv != "" {
		t.Fatalf("expected no Strategy-Failover-Provider when failover not triggered, got %s", failoverProv)
	}
}

func TestAIProxy_ModelMapping_ExactMatch(t *testing.T) {
	aliasName := "my-gpt4"
	providerName := "openai-mapped"
	realModel := "gpt-4o"

	ai_convert.SetKeyResource(providerName, &mockKeyResource{
		id: "mapped-key-1",
		driver: &mockConverterDriver{
			provider:  providerName,
			modelType: ai_convert.ModelTypeOpenAIChat,
		},
	})
	defer ai_convert.DelKeyResource(providerName, "mapped-key-1")

	// 注册模型映射策略
	cfg := &model_mapping.Config{
		Name:     "user-mapping-strat",
		Priority: 10,
		Filters: map[string][]string{
			"consumer": {"consumer-vip-1"},
			"model":    {aliasName},
		},
		Targets: []string{providerName + "/" + realModel},
	}
	handler, err := model_mapping.NewHandler(cfg)
	if err != nil {
		t.Fatalf("NewHandler err: %v", err)
	}
	model_mapping.SetStrategy(handler.Name(), handler, cfg.Filters)
	defer model_mapping.DelStrategy(handler.Name())

	expr, _ := jp.ParseString("model")
	exec := &executor{
		modelType:   ai_convert.ModelTypeOpenAIChat,
		modelIdFrom: "body",
		modelIdKey:  "model",
		bodyExpr:    expr,
		config:      "{}",
	}

	chain := &mockChain{
		fn: func(c eocontext.EoContext) error {
			httpCtx := c.(http_service.IHttpContext)
			httpCtx.Response().SetStatus(http.StatusOK, "OK")
			return nil
		},
	}

	// 构造请求体含有别名 {"model":"my-gpt4","messages":[]}
	rawReqBody := []byte(`{"model":"` + aliasName + `","messages":[{"role":"user","content":"hello"}]}`)
	ctx := newMockHttpContext("/v1/chat/completions", rawReqBody)
	ctx.SetLabel("consumer", "consumer-vip-1")

	err = exec.DoHttpFilter(ctx, chain)
	if err != nil {
		t.Fatalf("DoHttpFilter err: %v", err)
	}

	// 验证策略名响应头
	if ctx.Response().GetHeader("Strategy-Model-Mapping") != "user-mapping-strat" {
		t.Fatalf("expected Strategy-Model-Mapping header, got %s", ctx.Response().GetHeader("Strategy-Model-Mapping"))
	}
	// 验证最终实际解析出的供应商和模型
	if ctx.GetLabel("provider") != providerName {
		t.Fatalf("expected provider %s, got %s", providerName, ctx.GetLabel("provider"))
	}
	if ctx.GetLabel("model") != realModel {
		t.Fatalf("expected model %s, got %s", realModel, ctx.GetLabel("model"))
	}
	if ctx.Response().GetHeader("X-AI-Provider") != providerName {
		t.Fatalf("expected X-AI-Provider %s, got %s", providerName, ctx.Response().GetHeader("X-AI-Provider"))
	}
	if ctx.Response().GetHeader("X-AI-Model") != realModel {
		t.Fatalf("expected X-AI-Model %s, got %s", realModel, ctx.Response().GetHeader("X-AI-Model"))
	}

	// 核心业务红线验证：请求体中的 model 参数被精准改写为 realModel
	proxyBody, err := ctx.Proxy().Body().RawBody()
	if err != nil {
		t.Fatalf("read proxy body err: %v", err)
	}
	if !strings.Contains(string(proxyBody), `"`+realModel+`"`) {
		t.Fatalf("expected proxy body model rewritten to %s, got: %s", realModel, string(proxyBody))
	}
}

func TestAIProxy_ModelMapping_Failover(t *testing.T) {
	aliasName := "claude-pro"
	target1Provider := "anthropic-primary"
	target1Model := "claude-3-opus"
	target2Provider := "anthropic-backup"
	target2Model := "claude-3-5-sonnet"

	ai_convert.SetKeyResource(target1Provider, &mockKeyResource{
		id: "target1-key",
		driver: &mockConverterDriver{
			provider:  target1Provider,
			modelType: ai_convert.ModelTypeOpenAIChat,
		},
	})
	defer ai_convert.DelKeyResource(target1Provider, "target1-key")

	ai_convert.SetKeyResource(target2Provider, &mockKeyResource{
		id: "target2-key",
		driver: &mockConverterDriver{
			provider:  target2Provider,
			modelType: ai_convert.ModelTypeOpenAIChat,
		},
	})
	defer ai_convert.DelKeyResource(target2Provider, "target2-key")

	cfg := &model_mapping.Config{
		Name:     "claude-failover-map",
		Priority: 10,
		Filters: map[string][]string{
			"consumer": {"dev-team"},
			"model":    {aliasName},
		},
		Targets: []string{
			target1Provider + "/" + target1Model,
			target2Provider + "/" + target2Model,
		},
	}
	handler, err := model_mapping.NewHandler(cfg)
	if err != nil {
		t.Fatalf("NewHandler error: %v", err)
	}
	model_mapping.SetStrategy(handler.Name(), handler, cfg.Filters)
	defer model_mapping.DelStrategy(handler.Name())

	expr, _ := jp.ParseString("model")
	exec := &executor{
		modelType:   ai_convert.ModelTypeOpenAIChat,
		modelIdFrom: "body",
		modelIdKey:  "model",
		bodyExpr:    expr,
		config:      "{}",
	}

	callCount := 0
	chain := &mockChain{
		fn: func(c eocontext.EoContext) error {
			callCount++
			httpCtx := c.(http_service.IHttpContext)
			currProv := ai_convert.GetAIProvider(httpCtx)
			if currProv == target1Provider {
				httpCtx.Response().SetStatus(http.StatusServiceUnavailable, "Service Unavailable")
				httpCtx.Response().SetBody([]byte(`{"error":"primary down"}`))
				return errors.New("upstream service unavailable")
			}
			if currProv == target2Provider {
				httpCtx.Response().SetStatus(http.StatusOK, "OK")
				httpCtx.Response().SetBody([]byte(`{"result":"backup success"}`))
				return nil
			}
			return nil
		},
	}

	rawReqBody := []byte(`{"model":"` + aliasName + `"}`)
	ctx := newMockHttpContext("/v1/chat/completions", rawReqBody)
	ctx.SetLabel("consumer", "dev-team")

	err = exec.DoHttpFilter(ctx, chain)
	if err != nil {
		t.Fatalf("DoHttpFilter error: %v", err)
	}

	if callCount != 2 {
		t.Fatalf("expected 2 calls for failover, got %d", callCount)
	}

	// 验证最终生效的标签与响应头
	if ctx.GetLabel("provider") != target2Provider {
		t.Fatalf("expected provider to be %s, got %s", target2Provider, ctx.GetLabel("provider"))
	}
	if ctx.GetLabel("model") != target2Model {
		t.Fatalf("expected model to be %s, got %s", target2Model, ctx.GetLabel("model"))
	}
	if ctx.Response().GetHeader("Strategy-Model-Mapping-Failover-Provider") != target2Provider {
		t.Fatalf("expected failover header provider %s, got %s", target2Provider, ctx.Response().GetHeader("Strategy-Model-Mapping-Failover-Provider"))
	}
	if ctx.Response().GetHeader("Strategy-Model-Mapping-Failover-Model") != target2Model {
		t.Fatalf("expected failover header model %s, got %s", target2Model, ctx.Response().GetHeader("Strategy-Model-Mapping-Failover-Model"))
	}

	// 验证最终 proxy body 中的 model 是 target2Model
	proxyBody, err := ctx.Proxy().Body().RawBody()
	if err != nil {
		t.Fatalf("read proxy body err: %v", err)
	}
	if !strings.Contains(string(proxyBody), `"`+target2Model+`"`) {
		t.Fatalf("expected proxy body to have model %s, got %s", target2Model, string(proxyBody))
	}
}

func TestAIProxy_ModelMapping_FailoverPriorityOverPlatform(t *testing.T) {
	aliasName := "claude-coder"
	primaryProvider := "primary-anthropic"
	primaryModel := "claude-3-5-sonnet"
	consumerFailoverProvider := "consumer-anthropic-pro"
	consumerFailoverModel := "claude-3-opus"
	platformFailoverProvider := "platform-anthropic-backup"

	var primaryCalled, consumerFailoverCalled, platformFailoverCalled bool

	ai_convert.SetKeyResource(primaryProvider, &mockKeyResource{
		id: "primary-key",
		driver: &mockConverterDriver{
			provider:  primaryProvider,
			modelType: ai_convert.ModelTypeOpenAIChat,
			requestConvertFn: func(ctx eocontext.EoContext, extender map[string]interface{}) error {
				primaryCalled = true
				return nil
			},
		},
	})
	defer ai_convert.DelKeyResource(primaryProvider, "primary-key")

	ai_convert.SetKeyResource(consumerFailoverProvider, &mockKeyResource{
		id: "consumer-failover-key",
		driver: &mockConverterDriver{
			provider:  consumerFailoverProvider,
			modelType: ai_convert.ModelTypeOpenAIChat,
			requestConvertFn: func(ctx eocontext.EoContext, extender map[string]interface{}) error {
				consumerFailoverCalled = true
				return nil
			},
		},
	})
	defer ai_convert.DelKeyResource(consumerFailoverProvider, "consumer-failover-key")

	ai_convert.SetKeyResource(platformFailoverProvider, &mockKeyResource{
		id: "platform-failover-key",
		driver: &mockConverterDriver{
			provider:  platformFailoverProvider,
			modelType: ai_convert.ModelTypeOpenAIChat,
			requestConvertFn: func(ctx eocontext.EoContext, extender map[string]interface{}) error {
				platformFailoverCalled = true
				return nil
			},
		},
	})
	defer ai_convert.DelKeyResource(platformFailoverProvider, "platform-failover-key")

	// 1. 注册平台级供应商灾备策略：当 primaryProvider 失败时切换到 platformFailoverProvider
	platformFailoverCfg := &failover_strategy.Config{
		Name:     "platform-failover-strat",
		Priority: 5,
		Triggers: failover_strategy.TriggersConf{
			Failure: failover_strategy.TriggerFailureConf{
				Enabled: true,
			},
		},
		Filters: map[string][]string{
			"provider": {primaryProvider},
		},
		Providers: []*failover_strategy.ProviderConf{
			{Name: platformFailoverProvider},
		},
	}
	platformHandler, err := failover_strategy.NewHandler(platformFailoverCfg)
	if err != nil {
		t.Fatalf("failover_strategy.NewHandler err: %v", err)
	}
	failover_strategy.SetStrategy(platformHandler.Name(), platformHandler, platformFailoverCfg.Filters)
	defer failover_strategy.DelStrategy(platformHandler.Name())

	// 2. 注册消费者模型别名映射：配置了次选灾备目标 consumerFailoverProvider/consumerFailoverModel
	mappingCfg := &model_mapping.Config{
		Name:     "consumer-coder-mapping",
		Priority: 10,
		Filters: map[string][]string{
			"consumer": {"consumer-dev"},
			"model":    {aliasName},
		},
		Targets: []string{
			primaryProvider + "/" + primaryModel,
			consumerFailoverProvider + "/" + consumerFailoverModel,
		},
	}
	mappingHandler, err := model_mapping.NewHandler(mappingCfg)
	if err != nil {
		t.Fatalf("model_mapping.NewHandler err: %v", err)
	}
	model_mapping.SetStrategy(mappingHandler.Name(), mappingHandler, mappingCfg.Filters)
	defer model_mapping.DelStrategy(mappingHandler.Name())

	expr, _ := jp.ParseString("model")
	exec := &executor{
		modelType:   ai_convert.ModelTypeOpenAIChat,
		modelIdFrom: "body",
		modelIdKey:  "model",
		bodyExpr:    expr,
		config:      "{}",
	}

	// 模拟链路执行：Primary 遇到 500 报错，而消费者备选 Target 响应 200
	chain := &mockChain{
		fn: func(c eocontext.EoContext) error {
			httpCtx := c.(http_service.IHttpContext)
			currProv := ai_convert.GetAIProvider(httpCtx)
			if currProv == primaryProvider {
				httpCtx.Response().SetStatus(http.StatusInternalServerError, "Internal Server Error")
				httpCtx.Response().SetBody([]byte(`{"error":"upstream down"}`))
				return nil
			}
			if currProv == consumerFailoverProvider {
				httpCtx.Response().SetStatus(http.StatusOK, "OK")
				httpCtx.Response().SetBody([]byte(`{"result":"success from consumer backup"}`))
				return nil
			}
			if currProv == platformFailoverProvider {
				httpCtx.Response().SetStatus(http.StatusOK, "OK")
				return nil
			}
			return nil
		},
	}

	rawReqBody := []byte(`{"model":"` + aliasName + `"}`)
	ctx := newMockHttpContext("/v1/chat/completions", rawReqBody)
	ctx.SetLabel("consumer", "consumer-dev")

	err = exec.DoHttpFilter(ctx, chain)
	if err != nil {
		t.Fatalf("DoHttpFilter error: %v", err)
	}

	if !primaryCalled {
		t.Fatalf("expected primary provider to be attempted first")
	}
	if !consumerFailoverCalled {
		t.Fatalf("expected consumer failover target to be called, but wasn't")
	}
	if platformFailoverCalled {
		t.Fatalf("platform failover MUST NOT be called when consumer failover succeeded")
	}

	// 核心业务红线验证：切换后按实际接通的新模型精准计费（上下文中的 provider/model 必须变为消费者灾备目标）
	if ctx.GetLabel("provider") != consumerFailoverProvider {
		t.Fatalf("expected billing provider to be updated to %s, got %s", consumerFailoverProvider, ctx.GetLabel("provider"))
	}
	if ctx.GetLabel("model") != consumerFailoverModel {
		t.Fatalf("expected billing model to be updated to %s, got %s", consumerFailoverModel, ctx.GetLabel("model"))
	}
	if ctx.Response().GetHeader("Strategy-Model-Mapping-Failover-Provider") != consumerFailoverProvider {
		t.Fatalf("expected failover header provider %s, got %s", consumerFailoverProvider, ctx.Response().GetHeader("Strategy-Model-Mapping-Failover-Provider"))
	}
	if ctx.Response().GetHeader("Strategy-Model-Mapping-Failover-Model") != consumerFailoverModel {
		t.Fatalf("expected failover header model %s, got %s", consumerFailoverModel, ctx.Response().GetHeader("Strategy-Model-Mapping-Failover-Model"))
	}

	// 核心业务红线验证：请求体中的 model 被精准改写为消费者备选目标模型
	proxyBody, err := ctx.Proxy().Body().RawBody()
	if err != nil {
		t.Fatalf("read proxy body err: %v", err)
	}
	if !strings.Contains(string(proxyBody), `"`+consumerFailoverModel+`"`) {
		t.Fatalf("expected proxy body model rewritten to %s, got: %s", consumerFailoverModel, string(proxyBody))
	}
}

func TestAIProxy_ModelMapping_NoFallback_ShortCircuit(t *testing.T) {
	aliasName := "precise-alias"
	mappingCfg := &model_mapping.Config{
		Name:     "strict-mapping",
		Priority: 1,
		Filters: map[string][]string{
			"consumer": {"vip-only"},
			"model":    {aliasName},
		},
		Targets: []string{"openai/gpt-4"},
	}
	h, err := model_mapping.NewHandler(mappingCfg)
	if err != nil {
		t.Fatalf("NewHandler err: %v", err)
	}
	model_mapping.SetStrategy(h.Name(), h, mappingCfg.Filters)
	defer model_mapping.DelStrategy(h.Name())

	expr, _ := jp.ParseString("model")
	exec := &executor{
		modelType:   ai_convert.ModelTypeOpenAIChat,
		modelIdFrom: "body",
		modelIdKey:  "model",
		bodyExpr:    expr,
		config:      "{}",
	}

	chain := &mockChain{
		fn: func(c eocontext.EoContext) error {
			return nil
		},
	}

	// 请求缺失 consumer 标签（未认证），即便 model 命中也必须纯精确短路放行，不走该映射
	rawReqBody := []byte(`{"model":"` + aliasName + `"}`)
	ctx := newMockHttpContext("/v1/chat/completions", rawReqBody)
	// 未设置 consumer label

	err = exec.DoHttpFilter(ctx, chain)
	// 由于没有命中别名映射，也没有提供有效 provider（格式非 {provider}/{model} 且未在 balances 注册），应按标准逻辑报错
	if err == nil {
		t.Fatalf("expected error because extracted model has no provider and did not match mapping")
	}
	if ctx.Response().GetHeader("Strategy-Model-Mapping") != "" {
		t.Fatalf("Strategy-Model-Mapping should NOT be set when exact match fails")
	}
}
