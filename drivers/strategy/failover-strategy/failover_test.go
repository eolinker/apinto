package failover_strategy

import (
	"encoding/json"
	"net"
	"testing"
	"time"

	ai_convert "github.com/eolinker/apinto/ai-convert"
	http_context "github.com/eolinker/apinto/node/http-context"
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

func TestFailoverConfig(t *testing.T) {
	rawJSON := `{
		"name": "test-failover",
		"description": "A sample failover strategy",
		"driver": "failover",
		"template": "default",
		"filters": {
			"provider": [
				"openai"
			],
			"resource": [
				"openai/gpt-5"
			]
		},
		"trigger_type": "direct",
		"triggers": {
			"failure": {
				"enabled": true
			},
			"timeout": {
				"enabled": true,
				"timeout_seconds": 60
			}
		},
		"providers": [
			{
				"name": "azure-openai",
				"model": "gpt-5",
				"config": {
					"apikey": "secret-key",
					"base_url": "https://azure.openai.com/v1"
				}
			}
		]
	}`

	var conf Config
	if err := json.Unmarshal([]byte(rawJSON), &conf); err != nil {
		t.Fatalf("unmarshal config failed: %v", err)
	}

	if err := checkConfig(&conf); err != nil {
		t.Fatalf("checkConfig failed: %v", err)
	}

	if conf.Priority != 1 {
		t.Errorf("expected default priority 1, got %d", conf.Priority)
	}
}

func TestDirectFailoverHandler(t *testing.T) {
	conf := &Config{
		Name:        "test-failover",
		Description: "direct failover test",
		Priority:    10,
		TriggerType: TriggerTypeDirect,
		Filters: map[string][]string{
			"provider": {"openai"},
		},
		Providers: []*ProviderConf{
			{
				Name:  "qwen",
				Model: "qwen-max",
				Config: map[string]interface{}{
					"apikey": "qwen-secret",
				},
			},
		},
	}

	if err := checkConfig(conf); err != nil {
		t.Fatalf("checkConfig failed: %v", err)
	}

	handler, err := NewFailoverHandler(conf)
	if err != nil {
		t.Fatalf("NewFailoverHandler failed: %v", err)
	}

	ctx := newMockHttpContext()
	ctx.SetLabel("provider", "openai")
	ctx.SetLabel("model", "gpt-4")

	if !handler.Check(ctx) {
		t.Fatalf("handler.Check should return true for matching provider")
	}

	p, ok := handler.ApplyFailover(ctx, 0)
	if !ok || p == nil {
		t.Fatalf("ApplyFailover failed")
	}

	if ctx.GetLabel("provider") != "qwen" {
		t.Errorf("expected label provider 'qwen', got '%s'", ctx.GetLabel("provider"))
	}
	if ctx.GetLabel("model") != "qwen-max" {
		t.Errorf("expected label model 'qwen-max', got '%s'", ctx.GetLabel("model"))
	}
	if ai_convert.GetAIProvider(ctx) != "qwen" {
		t.Errorf("expected AI provider 'qwen', got '%s'", ai_convert.GetAIProvider(ctx))
	}
	if ctx.Response().GetHeader("Strategy-Failover") != "test-failover" {
		t.Errorf("expected Strategy-Failover header to be set")
	}
	if ctx.Response().GetHeader("X-Failover-Provider") != "qwen" {
		t.Errorf("expected X-Failover-Provider header to be 'qwen'")
	}
}

func TestConditionFailoverTrigger(t *testing.T) {
	conf := &Config{
		Name:        "condition-failover",
		Priority:    1,
		TriggerType: TriggerTypeCondition,
		Filters: map[string][]string{
			"resource": {"openai/gpt-4"},
		},
		Triggers: TriggersConf{
			Failure: TriggerFailureConf{
				Enabled:     true,
				StatusCodes: []int{500, 502, 503},
			},
			Timeout: TriggerTimeoutConf{
				Enabled:        true,
				TimeoutSeconds: 5,
			},
		},
		Providers: []*ProviderConf{
			{
				Name: "deepseek",
			},
		},
	}

	if err := checkConfig(conf); err != nil {
		t.Fatalf("checkConfig failed: %v", err)
	}

	handler, err := NewFailoverHandler(conf)
	if err != nil {
		t.Fatalf("NewFailoverHandler failed: %v", err)
	}

	ctx := newMockHttpContext()
	ctx.SetLabel("provider", "openai")
	ctx.SetLabel("model", "gpt-4")

	if !handler.Check(ctx) {
		t.Fatalf("handler.Check should match resource openai/gpt-4")
	}

	// 模拟正常响应
	ctx.Response().SetStatus(200, "OK")
	if handler.IsTriggerCondition(ctx, nil, 100*time.Millisecond) {
		t.Errorf("normal response should not trigger failover")
	}

	// 模拟 502 状态码触发
	ctx.Response().SetStatus(502, "Bad Gateway")
	if !handler.IsTriggerCondition(ctx, nil, 100*time.Millisecond) {
		t.Errorf("502 response should trigger failover")
	}

	// 模拟超时触发
	ctx.Response().SetStatus(200, "OK")
	if !handler.IsTriggerCondition(ctx, nil, 6*time.Second) {
		t.Errorf("cost >= 5s should trigger timeout failover")
	}
}

func TestActuatorSet(t *testing.T) {
	conf := &Config{
		Name:        "actuator-failover",
		Priority:    1,
		TriggerType: TriggerTypeDirect,
		Filters: map[string][]string{
			"provider": {"test-src"},
		},
		Providers: []*ProviderConf{
			{
				Name: "test-dst",
			},
		},
	}

	handler, err := NewFailoverHandler(conf)
	if err != nil {
		t.Fatalf("NewFailoverHandler failed: %v", err)
	}

	act := newtActuator()
	act.Set("test-id", handler)

	ctx := newMockHttpContext()
	ctx.SetLabel("provider", "test-src")

	err = act.Strategy(ctx, nil)
	if err != nil {
		t.Fatalf("act.Strategy failed: %v", err)
	}

	if ctx.GetLabel("provider") != "test-dst" {
		t.Errorf("expected provider to be switched to test-dst, got %s", ctx.GetLabel("provider"))
	}
}

func TestFailoverExtractor(t *testing.T) {
	// 策略 1 (Resource 粒度，Priority 50，数值大但属于特定 resource 策略)
	hResource, err := NewFailoverHandler(&Config{
		Name:        "strategy-resource-gpt5",
		Priority:    50,
		TriggerType: TriggerTypeDirect,
		Filters: map[string][]string{
			"resource": {"openai/gpt-5"},
		},
		Providers: []*ProviderConf{
			{Name: "azure-gpt5"},
		},
	})
	if err != nil {
		t.Fatalf("create resource handler failed: %v", err)
	}

	// 策略 2 (Provider 粒度，Priority 10，数值小优先级高，针对 openai 供应商)
	hProvider, err := NewFailoverHandler(&Config{
		Name:        "strategy-provider-openai",
		Priority:    10,
		TriggerType: TriggerTypeDirect,
		Filters: map[string][]string{
			"provider": {"openai"},
		},
		Providers: []*ProviderConf{
			{Name: "qwen-fallback"},
		},
	})
	if err != nil {
		t.Fatalf("create provider handler failed: %v", err)
	}

	// 策略 3 (全局兜底策略，Priority 100)
	hGeneral, err := NewFailoverHandler(&Config{
		Name:        "strategy-general",
		Priority:    100,
		TriggerType: TriggerTypeDirect,
		Filters:     map[string][]string{},
		Providers: []*ProviderConf{
			{Name: "global-fallback"},
		},
	})
	if err != nil {
		t.Fatalf("create general handler failed: %v", err)
	}

	extractor := NewExtractor()
	extractor.Set("id-resource", hResource)
	extractor.Set("id-provider", hProvider)
	extractor.Set("id-general", hGeneral)

	// 场景 1：请求的 resource 在筛选范围内 (openai/gpt-5)
	// 预期：命中 hResource (strategy-resource-gpt5)，即使 hProvider priority 为 10 也不考虑
	{
		ctx := newMockHttpContext()
		ctx.SetLabel("provider", "openai")
		ctx.SetLabel("model", "gpt-5")
		matched := extractor.Extract(ctx)
		if matched == nil {
			t.Fatalf("scenario 1: expected to match a handler, got nil")
		}
		if matched.Name() != "strategy-resource-gpt5" {
			t.Errorf("scenario 1: expected 'strategy-resource-gpt5', got '%s'", matched.Name())
		}
	}

	// 场景 2：请求的 resource 不在筛选范围内 (openai/gpt-4o)
	// 预期：resource 不在范围，回退去寻找 provider 与其匹配的策略，命中 hProvider
	{
		ctx := newMockHttpContext()
		ctx.SetLabel("provider", "openai")
		ctx.SetLabel("model", "gpt-4o")
		matched := extractor.Extract(ctx)
		if matched == nil {
			t.Fatalf("scenario 2: expected to match a handler, got nil")
		}
		if matched.Name() != "strategy-provider-openai" {
			t.Errorf("scenario 2: expected 'strategy-provider-openai', got '%s'", matched.Name())
		}
	}

	// 场景 3：请求的 resource 和 provider 都不匹配 (anthropic/claude-3-opus)
	// 预期：命中全局通用兜底策略 hGeneral
	{
		ctx := newMockHttpContext()
		ctx.SetLabel("provider", "anthropic")
		ctx.SetLabel("model", "claude-3-opus")
		matched := extractor.Extract(ctx)
		if matched == nil {
			t.Fatalf("scenario 3: expected to match general handler, got nil")
		}
		if matched.Name() != "strategy-general" {
			t.Errorf("scenario 3: expected 'strategy-general', got '%s'", matched.Name())
		}
	}

	// 场景 4：删除 resource 策略后，再次请求 openai/gpt-5
	// 预期：回退命中 provider 策略
	{
		extractor.Del("id-resource")
		ctx := newMockHttpContext()
		ctx.SetLabel("provider", "openai")
		ctx.SetLabel("model", "gpt-5")
		matched := extractor.Extract(ctx)
		if matched == nil {
			t.Fatalf("scenario 4: expected to match provider handler, got nil")
		}
		if matched.Name() != "strategy-provider-openai" {
			t.Errorf("scenario 4: expected 'strategy-provider-openai', got '%s'", matched.Name())
		}
	}
}
