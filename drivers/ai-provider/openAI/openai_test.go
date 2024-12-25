package openAI

import (
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"

	"github.com/eolinker/eosc/eocontext"

	ai_provider "github.com/eolinker/apinto/drivers/ai-provider"

	http_context "github.com/eolinker/apinto/node/http-context"

	"github.com/eolinker/apinto/convert"

	"github.com/valyala/fasthttp"
)

var (
	defaultConfig = `{
		"frequency_penalty": "",
		"max_tokens": 512,
		"presence_penalty": "",
		"response_format": "",
		"temperature": "",
		"top_p": ""
	}`
)

func validNormalFunc(ctx eocontext.EoContext) bool {
	fmt.Printf("input token: %d", ai_provider.GetAIModelInputToken(ctx))
	fmt.Printf("output token: %d", ai_provider.GetAIModelOutputToken(ctx))
	fmt.Printf("total token: %d", ai_provider.GetAIModelTotalToken(ctx))
	if ai_provider.GetAIModelInputToken(ctx) <= 0 {
		return false
	}
	if ai_provider.GetAIModelOutputToken(ctx) <= 0 {
		return false
	}
	return ai_provider.GetAIModelTotalToken(ctx) > 0
}

type tmpData struct {
	id     string
	name   string
	apiKey string
	want   string
	f      func(ctx eocontext.EoContext) bool
	body   []byte
}

// mock request body
var successBody = []byte(`{
	"messages": [
		{
			"content": "Hello, how can I help you?",
			"role": "assistant"
		}
	]
}`)

var failBody = []byte(`{
	"messages": [
		{
			"content": "Hello, how can I help you?",
			"role": "assistant"
		}
	],"variables":{}
}`)

func handle(data tmpData) error {
	cfg := &Config{
		APIKey: data.apiKey,
		Base:   "https://api.openai.com",
	}
	// Create the worker
	worker, err := Create("openai", "openai", cfg, nil)
	if err != nil {
		return fmt.Errorf("failed to create worker: %w", err)
	}

	// Validate worker implements IConverterDriver
	handler, ok := worker.(convert.IConverterDriver)
	if !ok {
		return fmt.Errorf("worker does not implement IConverterDriver")
	}

	// Load model
	model := "gpt-3.5-turbo"
	fn, has := handler.GetModel(model)
	if !has {
		return fmt.Errorf("model %s not found", model)
	}

	// Generate config
	extender, err := fn(defaultConfig)
	if err != nil {
		return fmt.Errorf("failed to generate config: %w", err)
	}
	requestBody := data.body
	if string(requestBody) == "" {
		requestBody = successBody
	}
	// Mock HTTP context
	ctx := createMockHttpContext("/xxx/xxx", nil, nil, requestBody)

	// Balance handler setup
	balanceHandler, err := ai_provider.NewBalanceHandler("test", cfg.Base, 30*time.Second)
	if err != nil {
		return fmt.Errorf("failed to create balance handler: %w", err)
	}

	// Execute the conversion process
	err = executeConverter(ctx, handler, model, extender, requestBody, balanceHandler)
	if err != nil {
		return fmt.Errorf("failed to execute conversion process: %w", err)
	}
	if ai_provider.GetAIStatus(ctx) != data.want {
		return fmt.Errorf("unexpected status: got %s, expected %s", ai_provider.GetAIStatus(ctx), data.want)
	}
	if data.f != nil && !data.f(ctx) {
		return fmt.Errorf("unexpected token status")
	}
	return nil
}

// TestSentTo tests the end-to-end execution of the OpenAI integration.
func TestSentTo(t *testing.T) {
	// Load API Key
	godotenv.Load(".env")
	testData := []tmpData{
		{
			id:     "openai",
			name:   "success",
			apiKey: os.Getenv("ValidKey"),
			want:   ai_provider.StatusNormal,
			f:      validNormalFunc,
		},
		{
			id:     "openai",
			name:   "invalid request",
			apiKey: os.Getenv("ValidKey"),
			want:   ai_provider.StatusInvalidRequest,
			body:   failBody,
		},
		{
			id:     "openai",
			name:   "invalid key",
			apiKey: os.Getenv("InvalidKey"),
			want:   ai_provider.StatusInvalid,
		},
		{
			id:     "openai",
			name:   "expired key",
			apiKey: os.Getenv("ExpiredKey"),
			want:   ai_provider.StatusInvalid,
		},
	}
	// Config setup
	for _, d := range testData {
		err := handle(d)
		if err != nil {
			t.Fatal(err)
		}
	}
}

// executeConverter handles the full flow of a conversion process.
func executeConverter(ctx *http_context.HttpContext, handler convert.IConverterDriver, model string, extender map[string]interface{}, body []byte, balanceHandler eocontext.BalanceHandler) error {
	// Get converter
	converter, has := handler.GetConverter(model)
	if !has {
		return fmt.Errorf("converter for model %s not found", model)
	}

	// Convert request
	if err := converter.RequestConvert(ctx, extender); err != nil {
		return fmt.Errorf("request conversion failed: %w", err)
	}

	// Select node via balance handler
	node, _, err := balanceHandler.Select(ctx)
	if err != nil {
		return fmt.Errorf("node selection failed: %w", err)
	}

	// Send request to the node
	if err := ctx.SendTo(balanceHandler.Scheme(), node, balanceHandler.TimeOut()); err != nil {
		return fmt.Errorf("failed to send request to node: %w", err)
	}

	// Convert response
	if err := converter.ResponseConvert(ctx); err != nil {
		return fmt.Errorf("response conversion failed: %w", err)
	}

	fmt.Printf("body: %s\n", string(ctx.Response().GetBody()))
	return nil
}

// createMockHttpContext creates a mock fasthttp.RequestCtx and wraps it with HttpContext.
func createMockHttpContext(rawURL string, headers map[string]string, query url.Values, body []byte) *http_context.HttpContext {
	req := fasthttp.AcquireRequest()
	u := fasthttp.AcquireURI()

	// Set request URI and path
	uri, _ := url.Parse(rawURL)
	u.SetPath(uri.Path)
	u.SetScheme(uri.Scheme)
	u.SetHost(uri.Host)
	u.SetQueryString(uri.RawQuery)
	req.SetURI(u)
	req.Header.SetMethod("POST")

	// Set headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	req.SetBody(body)

	// Create HttpContext
	return http_context.NewContext(&fasthttp.RequestCtx{
		Request:  *req,
		Response: fasthttp.Response{},
	}, 8099)
}
