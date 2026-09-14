package ai_proxy

import (
	"errors"
	"net"
	"net/http"
	"testing"
	"time"

	ai_convert "github.com/eolinker/apinto/ai-convert"
	context_label "github.com/eolinker/apinto/common/context-label"
	failover_strategy "github.com/eolinker/apinto/drivers/strategy/failover-strategy"
	http_context "github.com/eolinker/apinto/node/http-context"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
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

type mockKeyResource struct {
	id     string
	driver ai_convert.IConverterDriver
}

func (m *mockKeyResource) ID() string        { return m.id }
func (m *mockKeyResource) Health() bool      { return true }
func (m *mockKeyResource) Priority() int     { return 1 }
func (m *mockKeyResource) Up()               {}
func (m *mockKeyResource) Down()             {}
func (m *mockKeyResource) IsBreaker() bool   { return false }
func (m *mockKeyResource) Breaker()          {}
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
		Name:        "direct-failover-strat",
		Priority:    1,
		TriggerType: failover_strategy.TriggerTypeDirect,
		Filters: map[string][]string{
			"provider": {"failing-ai"},
		},
		Providers: []*failover_strategy.ProviderConf{
			{Name: backupProvider},
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
		Name:        "cond-failover-strat",
		Priority:    1,
		TriggerType: failover_strategy.TriggerTypeCondition,
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
		Name:        "timeout-failover-strat",
		Priority:    1,
		TriggerType: failover_strategy.TriggerTypeCondition,
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
		Name:        "disabled-timeout-strat",
		Priority:    1,
		TriggerType: failover_strategy.TriggerTypeCondition,
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
		Name:        "provider-strat",
		Priority:    10,
		TriggerType: failover_strategy.TriggerTypeDirect,
		Filters: map[string][]string{
			"provider": {"ai-svc"},
		},
		Providers: []*failover_strategy.ProviderConf{
			{Name: providerBackup},
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
		Name:        "resource-strat",
		Priority:    1,
		TriggerType: failover_strategy.TriggerTypeDirect,
		Filters: map[string][]string{
			"resource": {"ai-svc/special-model"},
		},
		Providers: []*failover_strategy.ProviderConf{
			{Name: resourceBackup},
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
