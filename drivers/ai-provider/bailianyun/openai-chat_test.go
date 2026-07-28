package bailianyun

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	ai_convert "github.com/eolinker/apinto/ai-convert"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/sashabaranov/go-openai"
)

// --- 1. 私有纯函数测试 ---

// toFloat64 将各种数值类型统一转为 float64，方便测试断言
func toFloat64(v interface{}) float64 {
	switch n := v.(type) {
	case int:
		return float64(n)
	case int32:
		return float64(n)
	case int64:
		return float64(n)
	case float32:
		return float64(n)
	case float64:
		return n
	default:
		return 0
	}
}

// approxEqual 比较 float64 是否近似相等（处理 float32 转 float64 的精度问题）
func approxEqual(a, b float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < 1e-5
}

func TestConvertOpenAIToDashScope(t *testing.T) {
	tests := []struct {
		name      string
		req       *openai.ChatCompletionRequest
		append    map[string]interface{}
		checkFunc func(t *testing.T, req *dashScopeRequest)
	}{
		{
			name: "basic text request",
			req: &openai.ChatCompletionRequest{
				Model: "qwen-turbo",
				Messages: []openai.ChatCompletionMessage{
					{Role: "system", Content: "You are a helpful assistant."},
					{Role: "user", Content: "Hello!"},
				},
				Temperature: 0.7,
				MaxTokens:   150,
			},
			append: map[string]interface{}{},
			checkFunc: func(t *testing.T, req *dashScopeRequest) {
				if req.Model != "qwen-turbo" {
					t.Errorf("expected model qwen-turbo, got %s", req.Model)
				}
				if len(req.Input.Messages) != 2 {
					t.Fatalf("expected 2 messages, got %d", len(req.Input.Messages))
				}
				if req.Input.Messages[0].Role != "system" {
					t.Errorf("expected first role system, got %s", req.Input.Messages[0].Role)
				}
				if req.Parameters["result_format"] != "message" {
					t.Errorf("expected result_format message, got %v", req.Parameters["result_format"])
				}
				if toFloat64(req.Parameters["max_tokens"]) != 150 {
					t.Errorf("expected max_tokens 150, got %v", req.Parameters["max_tokens"])
				}
				if !approxEqual(toFloat64(req.Parameters["temperature"]), 0.7) {
					t.Errorf("expected temperature 0.7, got %v", req.Parameters["temperature"])
				}
			},
		},
		{
			name: "stream request sets incremental output",
			req: &openai.ChatCompletionRequest{
				Model: "qwen-turbo",
				Messages: []openai.ChatCompletionMessage{
					{Role: "user", Content: "Hi"},
				},
				Stream: true,
			},
			append: map[string]interface{}{},
			checkFunc: func(t *testing.T, req *dashScopeRequest) {
				if req.Parameters["incremental_output"] != true {
					t.Errorf("expected incremental_output true, got %v", req.Parameters["incremental_output"])
				}
			},
		},
		{
			name: "MaxCompletionTokens fallback",
			req: &openai.ChatCompletionRequest{
				Model:               "qwen-max",
				Messages:            []openai.ChatCompletionMessage{{Role: "user", Content: "Hi"}},
				MaxCompletionTokens: 200,
			},
			append: map[string]interface{}{},
			checkFunc: func(t *testing.T, req *dashScopeRequest) {
				if toFloat64(req.Parameters["max_tokens"]) != 200 {
					t.Errorf("expected max_tokens 200, got %v", req.Parameters["max_tokens"])
				}
			},
		},
		{
			name: "append params merged",
			req: &openai.ChatCompletionRequest{
				Model:    "qwen-turbo",
				Messages: []openai.ChatCompletionMessage{{Role: "user", Content: "Hi"}},
			},
			append: map[string]interface{}{
				"custom_param": "custom_value",
			},
			checkFunc: func(t *testing.T, req *dashScopeRequest) {
				if req.Parameters["custom_param"] != "custom_value" {
					t.Errorf("expected custom_param custom_value, got %v", req.Parameters["custom_param"])
				}
			},
		},
		{
			name: "stop seed presence frequency penalty",
			req: &openai.ChatCompletionRequest{
				Model:            "qwen-turbo",
				Messages:         []openai.ChatCompletionMessage{{Role: "user", Content: "Hi"}},
				Stop:             []string{"END"},
				Seed:             new(int),
				PresencePenalty:  0.5,
				FrequencyPenalty: 0.3,
				TopP:             0.9,
			},
			append: map[string]interface{}{},
			checkFunc: func(t *testing.T, req *dashScopeRequest) {
				if !approxEqual(toFloat64(req.Parameters["top_p"]), 0.9) {
					t.Errorf("expected top_p 0.9, got %v", req.Parameters["top_p"])
				}
				if !approxEqual(toFloat64(req.Parameters["presence_penalty"]), 0.5) {
					t.Errorf("expected presence_penalty 0.5, got %v", req.Parameters["presence_penalty"])
				}
				if !approxEqual(toFloat64(req.Parameters["frequency_penalty"]), 0.3) {
					t.Errorf("expected frequency_penalty 0.3, got %v", req.Parameters["frequency_penalty"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertOpenAIToDashScope(tt.req, tt.append)
			tt.checkFunc(t, result)
		})
	}
}

func TestConvertOpenAIMessageToDashScope(t *testing.T) {
	t.Run("simple text content", func(t *testing.T) {
		msg := openai.ChatCompletionMessage{
			Role:    "user",
			Content: "Hello world",
			Name:    "test",
		}
		out := convertOpenAIMessageToDashScope(msg)
		if out.Role != "user" {
			t.Errorf("expected role user, got %s", out.Role)
		}
		if out.Name != "test" {
			t.Errorf("expected name test, got %s", out.Name)
		}
		if out.Content != "Hello world" {
			t.Errorf("expected content 'Hello world', got %v", out.Content)
		}
	})

	t.Run("multimodal content", func(t *testing.T) {
		msg := openai.ChatCompletionMessage{
			Role: "user",
			MultiContent: []openai.ChatMessagePart{
				{Type: openai.ChatMessagePartTypeText, Text: "Describe this"},
				{Type: openai.ChatMessagePartTypeImageURL, ImageURL: &openai.ChatMessageImageURL{URL: "https://example.com/img.png"}},
			},
		}
		out := convertOpenAIMessageToDashScope(msg)
		parts, ok := out.Content.([]map[string]interface{})
		if !ok {
			t.Fatalf("expected content to be []map[string]interface{}, got %T", out.Content)
		}
		if len(parts) != 2 {
			t.Fatalf("expected 2 parts, got %d", len(parts))
		}
		if parts[0]["text"] != "Describe this" {
			t.Errorf("expected first part text 'Describe this', got %v", parts[0]["text"])
		}
		if parts[1]["image"] != "https://example.com/img.png" {
			t.Errorf("expected second part image url, got %v", parts[1]["image"])
		}
	})

	t.Run("tool calls preserved", func(t *testing.T) {
		msg := openai.ChatCompletionMessage{
			Role: "assistant",
			ToolCalls: []openai.ToolCall{
				{ID: "call_1", Type: openai.ToolTypeFunction, Function: openai.FunctionCall{Name: "get_weather"}},
			},
			ToolCallID: "call_1",
		}
		out := convertOpenAIMessageToDashScope(msg)
		if len(out.ToolCalls) != 1 || out.ToolCalls[0].ID != "call_1" {
			t.Errorf("expected tool call call_1, got %v", out.ToolCalls)
		}
		if out.ToolCallID != "call_1" {
			t.Errorf("expected tool_call_id call_1, got %s", out.ToolCallID)
		}
	})
}

func TestDashScopeContentToString(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"string", "hello", "hello"},
		{"nil", nil, ""},
		{"string slice", []interface{}{}, ""},
		{
			name: "text parts",
			input: []interface{}{
				map[string]interface{}{"text": "Hello "},
				map[string]interface{}{"text": "World"},
			},
			expected: "Hello World",
		},
		{
			name: "mixed parts with non-text",
			input: []interface{}{
				map[string]interface{}{"image": "url"},
				map[string]interface{}{"text": "only text counts"},
			},
			expected: "only text counts",
		},
		{
			name:     "other type (number)",
			input:    42,
			expected: "42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := dashScopeContentToString(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestMapDashScopeFinishReason(t *testing.T) {
	tests := []struct {
		reason       string
		hasToolCalls bool
		expected     openai.FinishReason
	}{
		{"stop", false, openai.FinishReasonStop},
		{"null", false, openai.FinishReasonStop},
		{"", false, openai.FinishReasonStop},
		{"length", false, openai.FinishReasonLength},
		{"max_tokens", false, openai.FinishReasonLength},
		{"tool_calls", false, openai.FinishReasonToolCalls},
		{"content_filter", false, openai.FinishReasonContentFilter},
		{"stop", true, openai.FinishReasonToolCalls},
		{"unknown_reason", false, openai.FinishReasonStop},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%s/toolCalls=%v", tt.reason, tt.hasToolCalls)
		t.Run(name, func(t *testing.T) {
			result := mapDashScopeFinishReason(tt.reason, tt.hasToolCalls)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// --- 2. NewOpenAIChat 构造测试 ---

func TestNewOpenAIChat(t *testing.T) {
	t.Run("default base url", func(t *testing.T) {
		driver, err := NewOpenAIChat("bailian", "my-key", "", ai_convert.ModelTypeOpenAIChat, 5*time.Second)
		if err != nil {
			t.Fatalf("failed to create driver: %v", err)
		}
		chat, ok := driver.(*OpenAIChat)
		if !ok {
			t.Fatalf("driver is not *OpenAIChat")
		}
		if chat.apikey != "my-key" {
			t.Errorf("expected apikey my-key, got %s", chat.apikey)
		}
		if chat.provider != "bailian" {
			t.Errorf("expected provider bailian, got %s", chat.provider)
		}
		if chat.modelType != ai_convert.ModelTypeOpenAIChat {
			t.Errorf("expected modelType OpenAIChat, got %v", chat.modelType)
		}
		expectedPath := "/api/v1" + openaiChatPath
		if chat.path != expectedPath {
			t.Errorf("expected imagePath %s, got %s", expectedPath, chat.path)
		}
	})

	t.Run("custom base url with sub imagePath", func(t *testing.T) {
		driver, err := NewOpenAIChat("bailian", "my-key", "https://api.example.com/custom-prefix", ai_convert.ModelTypeOpenAIChat, 5*time.Second)
		if err != nil {
			t.Fatalf("failed to create driver: %v", err)
		}
		chat := driver.(*OpenAIChat)
		if !strings.Contains(chat.path, "/custom-prefix") {
			t.Errorf("expected imagePath to contain /custom-prefix, got %s", chat.path)
		}
		if !strings.HasSuffix(chat.path, openaiChatPath) {
			t.Errorf("expected imagePath to end with %s, got %s", openaiChatPath, chat.path)
		}
	})

	t.Run("trailing slash in base url", func(t *testing.T) {
		driver, err := NewOpenAIChat("bailian", "my-key", "https://api.example.com/", ai_convert.ModelTypeOpenAIChat, 5*time.Second)
		if err != nil {
			t.Fatalf("failed to create driver: %v", err)
		}
		chat := driver.(*OpenAIChat)
		expected := openaiChatPath
		if !strings.HasSuffix(chat.path, expected) {
			t.Errorf("expected imagePath to end with %s, got %s", expected, chat.path)
		}
	})

	t.Run("invalid base url", func(t *testing.T) {
		_, err := NewOpenAIChat("bailian", "my-key", "::invalid-url::", ai_convert.ModelTypeOpenAIChat, 5*time.Second)
		if err == nil {
			t.Errorf("expected error for invalid URL, got nil")
		}
	})
}

// --- 3. Mock 结构体定义 ---

type mockHeader struct {
	http_service.IHeaderWriter
	headers map[string]string
}

func (m *mockHeader) SetHeader(key, value string) {
	if m.headers == nil {
		m.headers = make(map[string]string)
	}
	m.headers[key] = value
}

func (m *mockHeader) DelHeader(key string) {
	if m.headers != nil {
		delete(m.headers, key)
	}
}

func (m *mockHeader) Get(key string) string {
	if m.headers == nil {
		return ""
	}
	return m.headers[key]
}

type mockBody struct {
	http_service.IBodyDataWriter
	contentType string
	body        []byte
}

func (m *mockBody) RawBody() ([]byte, error) {
	return m.body, nil
}

func (m *mockBody) SetRaw(contentType string, body []byte) {
	m.contentType = contentType
	m.body = body
}

type mockURI struct {
	http_service.IURIWriter
	path  string
	query url.Values
}

func (m *mockURI) SetPath(path string) {
	m.path = path
}

func (m *mockURI) SetQuery(key, value string) {
	if m.query == nil {
		m.query = make(url.Values)
	}
	m.query.Set(key, value)
}

type mockRequest struct {
	http_service.IRequest
	header         *mockHeader
	body           *mockBody
	uri            *mockURI
	streamHandlers []http_service.StreamFunc
}

func (m *mockRequest) Header() http_service.IHeaderWriter {
	return m.header
}

func (m *mockRequest) Body() http_service.IBodyDataWriter {
	return m.body
}

func (m *mockRequest) URI() http_service.IURIWriter {
	return m.uri
}

func (m *mockRequest) AppendStreamBodyHandle(handler http_service.StreamFunc) {
	m.streamHandlers = append(m.streamHandlers, handler)
}

type mockResponse struct {
	http_service.IResponse
	statusCode int
	body       []byte
	headers    map[string]string
}

func (m *mockResponse) StatusCode() int {
	return m.statusCode
}

func (m *mockResponse) GetBody() []byte {
	return m.body
}

func (m *mockResponse) SetBody(body []byte) {
	m.body = body
}

func (m *mockResponse) SetHeader(key, value string) {
	if m.headers == nil {
		m.headers = make(map[string]string)
	}
	m.headers[key] = value
}

func (m *mockResponse) Headers() http.Header {
	h := make(http.Header)
	for k, v := range m.headers {
		h.Set(k, v)
	}
	return h
}

type mockHttpContext struct {
	http_service.IHttpContext
	proxy          *mockRequest
	response       *mockResponse
	values         map[interface{}]interface{}
	labels         map[string]string
	requestId      string
	balanceHandler eocontext.BalanceHandler
}

func (m *mockHttpContext) Proxy() http_service.IRequest {
	return m.proxy
}

func (m *mockHttpContext) Response() http_service.IResponse {
	return m.response
}

func (m *mockHttpContext) Value(key interface{}) interface{} {
	if m.values == nil {
		return nil
	}
	return m.values[key]
}

func (m *mockHttpContext) WithValue(key, val interface{}) {
	if m.values == nil {
		m.values = make(map[interface{}]interface{})
	}
	m.values[key] = val
}

func (m *mockHttpContext) SetLabel(key, val string) {
	if m.labels == nil {
		m.labels = make(map[string]string)
	}
	m.labels[key] = val
}

func (m *mockHttpContext) GetLabel(key string) string {
	if m.labels == nil {
		return ""
	}
	return m.labels[key]
}

func (m *mockHttpContext) RequestId() string {
	return m.requestId
}

func (m *mockHttpContext) SetBalance(handler eocontext.BalanceHandler) {
	m.balanceHandler = handler
}

func (m *mockHttpContext) Assert(i interface{}) error {
	if v, ok := i.(*http_service.IHttpContext); ok {
		*v = m
		return nil
	}
	return fmt.Errorf("not http context")
}

// --- 4. RequestConvert / ResponseConvert 集成测试 ---

func TestRequestConvert(t *testing.T) {
	chat, err := NewOpenAIChat("bailian", "api-key-xyz", "https://api.example.com", ai_convert.ModelTypeOpenAIChat, 5*time.Second)
	if err != nil {
		t.Fatalf("failed to create driver: %v", err)
	}

	chatReq := openai.ChatCompletionRequest{
		Model: "qwen-turbo",
		Messages: []openai.ChatCompletionMessage{
			{Role: "system", Content: "You are an AI assistant."},
			{Role: "user", Content: "Hello!"},
		},
		Temperature: 0.7,
		MaxTokens:   150,
	}
	reqBytes, _ := json.Marshal(chatReq)

	mCtx := &mockHttpContext{
		proxy: &mockRequest{
			header: &mockHeader{},
			body:   &mockBody{body: reqBytes},
			uri:    &mockURI{},
		},
		requestId: "req-123456",
	}
	ai_convert.SetAIModel(mCtx, "qwen-turbo")

	err = chat.RequestConvert(mCtx, map[string]interface{}{})
	if err != nil {
		t.Fatalf("RequestConvert failed: %v", err)
	}

	// 验证 Authorization header
	authHeader := mCtx.proxy.header.Get("Authorization")
	if authHeader != "Bearer api-key-xyz" {
		t.Errorf("expected Authorization 'Bearer api-key-xyz', got %s", authHeader)
	}

	// 验证路径设置
	if !strings.HasSuffix(mCtx.proxy.uri.path, openaiChatPath) {
		t.Errorf("expected imagePath to end with %s, got %s", openaiChatPath, mCtx.proxy.uri.path)
	}

	// 验证 body 被转换为 DashScope 格式
	var dashReq dashScopeRequest
	err = json.Unmarshal(mCtx.proxy.body.body, &dashReq)
	if err != nil {
		t.Fatalf("failed to unmarshal converted body: %v", err)
	}
	if dashReq.Model != "qwen-turbo" {
		t.Errorf("expected model qwen-turbo, got %s", dashReq.Model)
	}
	if len(dashReq.Input.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(dashReq.Input.Messages))
	}
	if dashReq.Input.Messages[0].Role != "system" {
		t.Errorf("expected first role system, got %s", dashReq.Input.Messages[0].Role)
	}
	if dashReq.Parameters["result_format"] != "message" {
		t.Errorf("expected result_format message, got %v", dashReq.Parameters["result_format"])
	}
	if dashReq.Parameters["max_tokens"].(float64) != 150 {
		t.Errorf("expected max_tokens 150, got %v", dashReq.Parameters["max_tokens"])
	}

	// 验证 content-type
	if mCtx.proxy.body.contentType != "application/json" {
		t.Errorf("expected content-type application/json, got %s", mCtx.proxy.body.contentType)
	}
}

func TestRequestConvertStream(t *testing.T) {
	chat, _ := NewOpenAIChat("bailian", "api-key-xyz", "https://api.example.com", ai_convert.ModelTypeOpenAIChat, 5*time.Second)

	chatReq := openai.ChatCompletionRequest{
		Model:    "qwen-turbo",
		Messages: []openai.ChatCompletionMessage{{Role: "user", Content: "Hello!"}},
		Stream:   true,
	}
	reqBytes, _ := json.Marshal(chatReq)

	mCtx := &mockHttpContext{
		proxy: &mockRequest{
			header: &mockHeader{},
			body:   &mockBody{body: reqBytes},
			uri:    &mockURI{},
		},
		requestId: "req-stream-789",
	}
	ai_convert.SetAIModel(mCtx, "qwen-turbo")

	err := chat.RequestConvert(mCtx, map[string]interface{}{})
	if err != nil {
		t.Fatalf("RequestConvert failed: %v", err)
	}

	// 验证 stream header 设置
	sseHeader := mCtx.proxy.header.Get("X-DashScope-SSE")
	if sseHeader != "enable" {
		t.Errorf("expected X-DashScope-SSE enable, got %s", sseHeader)
	}

	// 验证注册了 stream handler
	if len(mCtx.proxy.streamHandlers) != 1 {
		t.Fatalf("expected 1 stream handler, got %d", len(mCtx.proxy.streamHandlers))
	}

	// 验证 incremental_output 参数
	var dashReq dashScopeRequest
	_ = json.Unmarshal(mCtx.proxy.body.body, &dashReq)
	if dashReq.Parameters["incremental_output"] != true {
		t.Errorf("expected incremental_output true, got %v", dashReq.Parameters["incremental_output"])
	}

	// 验证 response-content-type label
	ct := mCtx.GetLabel("response-content-type")
	if ct != "text/event-stream" {
		t.Errorf("expected response-content-type text/event-stream, got %s", ct)
	}
}

func TestRequestConvertEmptyModel(t *testing.T) {
	chat, _ := NewOpenAIChat("bailian", "api-key-xyz", "https://api.example.com", ai_convert.ModelTypeOpenAIChat, 5*time.Second)

	chatReq := openai.ChatCompletionRequest{
		Model:    "",
		Messages: []openai.ChatCompletionMessage{{Role: "user", Content: "Hello!"}},
	}
	reqBytes, _ := json.Marshal(chatReq)

	mCtx := &mockHttpContext{
		proxy: &mockRequest{
			header: &mockHeader{},
			body:   &mockBody{body: reqBytes},
			uri:    &mockURI{},
		},
		requestId: "req-model-empty",
	}
	ai_convert.SetAIModel(mCtx, "fallback-model")

	err := chat.RequestConvert(mCtx, map[string]interface{}{})
	if err != nil {
		t.Fatalf("RequestConvert failed: %v", err)
	}

	var dashReq dashScopeRequest
	_ = json.Unmarshal(mCtx.proxy.body.body, &dashReq)
	if dashReq.Model != "fallback-model" {
		t.Errorf("expected model fallback-model from context, got %s", dashReq.Model)
	}
}

func TestResponseConvert(t *testing.T) {
	chat, _ := NewOpenAIChat("bailian", "api-key-xyz", "https://api.example.com", ai_convert.ModelTypeOpenAIChat, 5*time.Second)

	dashResp := dashScopeResponse{
		Output: dashScopeOutput{
			Choices: []dashScopeChoice{
				{
					FinishReason: "stop",
					Message: dashScopeMessage{
						Role:    "assistant",
						Content: "Hello from DashScope!",
					},
				},
			},
		},
		Usage: dashScopeUsage{
			InputTokens:  10,
			OutputTokens: 20,
			TotalTokens:  30,
		},
		RequestID: "req-abc-123",
	}
	respBytes, _ := json.Marshal(dashResp)

	mCtx := &mockHttpContext{
		response: &mockResponse{
			statusCode: 200,
			body:       respBytes,
			headers:    map[string]string{"content-encoding": "utf-8"},
		},
		requestId: "req-123",
	}
	ai_convert.SetAIModel(mCtx, "qwen-turbo")

	err := chat.ResponseConvert(mCtx)
	if err != nil {
		t.Fatalf("ResponseConvert failed: %v", err)
	}

	// 验证 Response 被转换为 OpenAI 格式
	var openaiResp openai.ChatCompletionResponse
	err = json.Unmarshal(mCtx.response.body, &openaiResp)
	if err != nil {
		t.Fatalf("failed to unmarshal converted body: %v", err)
	}

	if openaiResp.Model != "qwen-turbo" {
		t.Errorf("expected model qwen-turbo, got %s", openaiResp.Model)
	}
	if !strings.HasPrefix(openaiResp.ID, "chatcmpl-") {
		t.Errorf("expected ID to start with chatcmpl-, got %s", openaiResp.ID)
	}
	if openaiResp.Object != "chat.completion" {
		t.Errorf("expected object chat.completion, got %s", openaiResp.Object)
	}
	if len(openaiResp.Choices) != 1 {
		t.Fatalf("expected 1 choice, got %d", len(openaiResp.Choices))
	}
	if openaiResp.Choices[0].Message.Content != "Hello from DashScope!" {
		t.Errorf("expected content 'Hello from DashScope!', got %s", openaiResp.Choices[0].Message.Content)
	}
	if openaiResp.Choices[0].FinishReason != openai.FinishReasonStop {
		t.Errorf("expected finish reason stop, got %v", openaiResp.Choices[0].FinishReason)
	}
	if openaiResp.Usage.PromptTokens != 10 || openaiResp.Usage.CompletionTokens != 20 || openaiResp.Usage.TotalTokens != 30 {
		t.Errorf("usage mismatch: prompt=%d, completion=%d, total=%d",
			openaiResp.Usage.PromptTokens, openaiResp.Usage.CompletionTokens, openaiResp.Usage.TotalTokens)
	}

	// 验证上下文中的 Token 标签
	if ai_convert.GetAIModelInputToken(mCtx) != 10 {
		t.Errorf("expected input token 10, got %d", ai_convert.GetAIModelInputToken(mCtx))
	}
	if ai_convert.GetAIModelOutputToken(mCtx) != 20 {
		t.Errorf("expected output token 20, got %d", ai_convert.GetAIModelOutputToken(mCtx))
	}
	if ai_convert.GetAIModelTotalToken(mCtx) != 30 {
		t.Errorf("expected total token 30, got %d", ai_convert.GetAIModelTotalToken(mCtx))
	}

	// 验证 content-encoding 被设置为 utf-8
	if mCtx.response.headers["content-encoding"] != "utf-8" {
		t.Errorf("expected content-encoding utf-8, got %s", mCtx.response.headers["content-encoding"])
	}
}

func TestResponseConvertTextFallback(t *testing.T) {
	chat, _ := NewOpenAIChat("bailian", "api-key-xyz", "", ai_convert.ModelTypeOpenAIChat, 5*time.Second)

	// 当 Choices 为空时，使用 Output.Text 作为内容
	dashResp := dashScopeResponse{
		Output: dashScopeOutput{
			Text:         "This is text output.",
			FinishReason: "stop",
		},
		Usage: dashScopeUsage{
			InputTokens:  5,
			OutputTokens: 5,
			TotalTokens:  10,
		},
	}
	respBytes, _ := json.Marshal(dashResp)

	mCtx := &mockHttpContext{
		response: &mockResponse{
			statusCode: 200,
			body:       respBytes,
		},
		requestId: "req-text-456",
	}
	ai_convert.SetAIModel(mCtx, "qwen-turbo")

	err := chat.ResponseConvert(mCtx)
	if err != nil {
		t.Fatalf("ResponseConvert failed: %v", err)
	}

	var openaiResp openai.ChatCompletionResponse
	_ = json.Unmarshal(mCtx.response.body, &openaiResp)
	if len(openaiResp.Choices) != 1 {
		t.Fatalf("expected 1 choice, got %d", len(openaiResp.Choices))
	}
	if openaiResp.Choices[0].Message.Content != "This is text output." {
		t.Errorf("expected content 'This is text output.', got %s", openaiResp.Choices[0].Message.Content)
	}
	if openaiResp.Choices[0].Message.Role != "assistant" {
		t.Errorf("expected role assistant, got %s", openaiResp.Choices[0].Message.Role)
	}
}

func TestResponseConvertError(t *testing.T) {
	chat, _ := NewOpenAIChat("bailian", "api-key-xyz", "", ai_convert.ModelTypeOpenAIChat, 5*time.Second)

	dashErr := dashScopeResponse{
		Code:    "InvalidApiKey",
		Message: "Invalid API key provided.",
	}
	respBytes, _ := json.Marshal(dashErr)

	mCtx := &mockHttpContext{
		response: &mockResponse{
			statusCode: 401,
			body:       respBytes,
		},
		requestId: "req-err-789",
	}

	err := chat.ResponseConvert(mCtx)
	if err != nil {
		t.Fatalf("ResponseConvert should not return error for non-200 status: %v", err)
	}

	var openaiErr struct {
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Code    string `json:"code"`
		} `json:"error"`
	}
	err = json.Unmarshal(mCtx.response.body, &openaiErr)
	if err != nil {
		t.Fatalf("failed to unmarshal converted error body: %v", err)
	}
	if openaiErr.Error.Message != "Invalid API key provided." {
		t.Errorf("expected message 'Invalid API key provided.', got %s", openaiErr.Error.Message)
	}
	if openaiErr.Error.Code != "InvalidApiKey" {
		t.Errorf("expected code 'InvalidApiKey', got %s", openaiErr.Error.Code)
	}
}

func TestConvertErrorResponse(t *testing.T) {
	chat, _ := NewOpenAIChat("bailian", "api-key", "", ai_convert.ModelTypeOpenAIChat, 5*time.Second)
	chatImpl := chat.(*OpenAIChat)

	t.Run("valid error response", func(t *testing.T) {
		dashErr := dashScopeResponse{
			Code:    "Throttling",
			Message: "Request was throttled.",
		}
		body, _ := json.Marshal(dashErr)
		mCtx := &mockHttpContext{
			response: &mockResponse{statusCode: 429, body: body},
		}
		chatImpl.convertErrorResponse(mCtx, body)

		var openaiErr struct {
			Error struct {
				Message string `json:"message"`
				Type    string `json:"type"`
				Code    string `json:"code"`
			} `json:"error"`
		}
		err := json.Unmarshal(mCtx.response.body, &openaiErr)
		if err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}
		if openaiErr.Error.Message != "Request was throttled." {
			t.Errorf("expected message 'Request was throttled.', got %s", openaiErr.Error.Message)
		}
		if openaiErr.Error.Code != "Throttling" {
			t.Errorf("expected code 'Throttling', got %s", openaiErr.Error.Code)
		}
	})

	t.Run("empty message returns early", func(t *testing.T) {
		body := []byte(`{"code":"SomeCode"}`)
		mCtx := &mockHttpContext{
			response: &mockResponse{
				statusCode: 400,
				body:       body,
			},
		}
		chatImpl.convertErrorResponse(mCtx, body)
		// body 应当未被修改，因为 Message 为空
		if string(mCtx.response.body) != string(body) {
			t.Errorf("expected body to be unchanged when message is empty")
		}
	})

	t.Run("invalid json returns early", func(t *testing.T) {
		body := []byte("{invalid json}")
		mCtx := &mockHttpContext{
			response: &mockResponse{
				statusCode: 500,
				body:       body,
			},
		}
		chatImpl.convertErrorResponse(mCtx, body)
		if string(mCtx.response.body) != string(body) {
			t.Errorf("expected body to be unchanged for invalid json")
		}
	})
}

// --- 5. Stream Handler 集成测试 ---

func TestStreamHandler(t *testing.T) {
	chat, _ := NewOpenAIChat("bailian", "api-key", "", ai_convert.ModelTypeOpenAIChat, 5*time.Second)
	chatImpl := chat.(*OpenAIChat)

	mCtx := &mockHttpContext{
		requestId: "req-stream-456",
		labels:    map[string]string{},
	}
	ai_convert.SetAIModel(mCtx, "qwen-turbo")

	// 构造 SSE 数据
	chunk1 := []byte("data: {\"output\":{\"choices\":[{\"finish_reason\":\"\",\"message\":{\"role\":\"assistant\",\"content\":\"Hello \"}}]},\"request_id\":\"req-1\"}\n")
	chunk2 := []byte("data: {\"output\":{\"choices\":[{\"finish_reason\":\"stop\",\"message\":{\"role\":\"assistant\",\"content\":\"World\"}}]},\"usage\":{\"input_tokens\":5,\"output_tokens\":5,\"total_tokens\":10},\"request_id\":\"req-1\"}\n\n")
	chunk3 := []byte("data: [DONE]\n\n")

	// 测试第一段
	out1, err := chatImpl.streamHandler(mCtx, chunk1)
	if err != nil {
		t.Fatalf("streamHandler failed on chunk1: %v", err)
	}
	if len(out1) == 0 {
		t.Fatalf("expected output for chunk1, got empty")
	}
	if !strings.HasPrefix(string(out1), "data:") {
		t.Errorf("expected output to start with 'data:', got %s", string(out1))
	}

	// 解析第一个 chunk 的内容
	lines := strings.Split(strings.TrimSpace(string(out1)), "\n\n")
	var streamResp1 openai.ChatCompletionStreamResponse
	cleanData1 := strings.TrimPrefix(lines[0], "data: ")
	err = json.Unmarshal([]byte(cleanData1), &streamResp1)
	if err != nil {
		t.Fatalf("failed to unmarshal stream output1: %v, data: %s", err, cleanData1)
	}
	if streamResp1.Choices[0].Delta.Content != "Hello " {
		t.Errorf("expected content 'Hello ', got %s", streamResp1.Choices[0].Delta.Content)
	}
	if streamResp1.Choices[0].Delta.Role != "assistant" {
		t.Errorf("expected role assistant, got %s", streamResp1.Choices[0].Delta.Role)
	}

	// 测试第二段（含 usage）
	out2, err := chatImpl.streamHandler(mCtx, chunk2)
	if err != nil {
		t.Fatalf("streamHandler failed on chunk2: %v", err)
	}
	if len(out2) == 0 {
		t.Fatalf("expected output for chunk2, got empty")
	}

	// 验证 token 计数
	if ai_convert.GetAIModelInputToken(mCtx) != 5 {
		t.Errorf("expected input tokens 5, got %d", ai_convert.GetAIModelInputToken(mCtx))
	}
	if ai_convert.GetAIModelTotalToken(mCtx) != 10 {
		t.Errorf("expected total tokens 10, got %d", ai_convert.GetAIModelTotalToken(mCtx))
	}

	// 测试 [DONE]
	out3, err := chatImpl.streamHandler(mCtx, chunk3)
	if err != nil {
		t.Fatalf("streamHandler failed on chunk3: %v", err)
	}
	if !strings.Contains(string(out3), "data: [DONE]") {
		t.Errorf("expected output to contain 'data: [DONE]', got %s", string(out3))
	}
}

func TestStreamHandlerPartialChunk(t *testing.T) {
	chat, _ := NewOpenAIChat("bailian", "api-key", "", ai_convert.ModelTypeOpenAIChat, 5*time.Second)
	chatImpl := chat.(*OpenAIChat)

	mCtx := &mockHttpContext{
		requestId: "req-partial-001",
		labels:    map[string]string{bailianStreamRemainLabel: ""},
	}
	ai_convert.SetAIModel(mCtx, "qwen-turbo")

	// 测试不完整的 chunk（无换行结尾），应当全部缓存到 remain label
	partial := []byte(`data: {"output":{"choices":[{"finish_reason":"","message":{"role":"assistant","content":"partial"}}]},"request_id":"req-1"}`)
	out, err := chatImpl.streamHandler(mCtx, partial)
	if err != nil {
		t.Fatalf("streamHandler failed: %v", err)
	}
	// 没有完整换行，应该返回空，且数据缓存在 label 中
	if len(out) != 0 {
		t.Errorf("expected empty output for partial chunk, got %d bytes", len(out))
	}
	remain := mCtx.GetLabel(bailianStreamRemainLabel)
	if remain == "" {
		t.Errorf("expected remain label to be set for partial chunk")
	}
	if !strings.Contains(remain, "partial") {
		t.Errorf("expected remain to contain 'partial', got %s", remain)
	}

	// 补全换行，使数据完整可解析
	complete := []byte("\n\n")
	out2, err := chatImpl.streamHandler(mCtx, complete)
	if err != nil {
		t.Fatalf("streamHandler failed on completion: %v", err)
	}
	if len(out2) == 0 {
		t.Fatalf("expected output after completing partial chunk")
	}

	// 解析输出
	cleanData := strings.TrimPrefix(strings.TrimSuffix(string(out2), "\n\n"), "data: ")
	var streamResp openai.ChatCompletionStreamResponse
	err = json.Unmarshal([]byte(cleanData), &streamResp)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v, data: %s", err, cleanData)
	}
	if streamResp.Choices[0].Delta.Content != "partial" {
		t.Errorf("expected content 'partial', got %s", streamResp.Choices[0].Delta.Content)
	}

	// remain 应当被清空
	remain = mCtx.GetLabel(bailianStreamRemainLabel)
	if remain != "" {
		t.Errorf("expected remain label to be empty after completion, got %s", remain)
	}
}

func TestStreamHandlerInvalidJSON(t *testing.T) {
	chat, _ := NewOpenAIChat("bailian", "api-key", "", ai_convert.ModelTypeOpenAIChat, 5*time.Second)
	chatImpl := chat.(*OpenAIChat)

	mCtx := &mockHttpContext{
		requestId: "req-invalid-001",
		labels:    map[string]string{},
	}
	ai_convert.SetAIModel(mCtx, "qwen-turbo")

	// 无效 JSON 的 data 行应当被跳过，不产生输出
	invalid := []byte("data: {invalid json}\n\n")
	out, err := chatImpl.streamHandler(mCtx, invalid)
	if err != nil {
		t.Fatalf("streamHandler failed: %v", err)
	}
	if len(out) != 0 {
		t.Errorf("expected empty output for invalid json, got %d bytes", len(out))
	}

	// 非 data: 开头的行也应被跳过
	nonData := []byte(": comment line\n\nevent: ping\n\n")
	out2, err := chatImpl.streamHandler(mCtx, nonData)
	if err != nil {
		t.Fatalf("streamHandler failed: %v", err)
	}
	if len(out2) != 0 {
		t.Errorf("expected empty output for non-data lines, got %d bytes", len(out2))
	}
}

func TestStreamHandlerMultipleChunks(t *testing.T) {
	chat, _ := NewOpenAIChat("bailian", "api-key", "", ai_convert.ModelTypeOpenAIChat, 5*time.Second)
	chatImpl := chat.(*OpenAIChat)

	mCtx := &mockHttpContext{
		requestId: "req-multi-001",
		labels:    map[string]string{},
	}
	ai_convert.SetAIModel(mCtx, "qwen-turbo")

	// 一次传入多个 SSE event
	multi := []byte("data: {\"output\":{\"choices\":[{\"finish_reason\":\"\",\"message\":{\"role\":\"assistant\",\"content\":\"A\"}}]},\"request_id\":\"req-1\"}\n\ndata: {\"output\":{\"choices\":[{\"finish_reason\":\"\",\"message\":{\"role\":\"assistant\",\"content\":\"B\"}}]},\"request_id\":\"req-1\"}\n\n")

	out, err := chatImpl.streamHandler(mCtx, multi)
	if err != nil {
		t.Fatalf("streamHandler failed: %v", err)
	}

	// 应当输出两个 SSE event
	events := strings.Split(strings.TrimSpace(string(out)), "\n\n")
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d: %s", len(events), string(out))
	}

	contents := []string{}
	for _, event := range events {
		cleanData := strings.TrimPrefix(event, "data: ")
		var streamResp openai.ChatCompletionStreamResponse
		if err := json.Unmarshal([]byte(cleanData), &streamResp); err != nil {
			t.Fatalf("failed to unmarshal: %v, data: %s", err, cleanData)
		}
		contents = append(contents, streamResp.Choices[0].Delta.Content)
	}
	if contents[0] != "A" || contents[1] != "B" {
		t.Errorf("expected contents ['A','B'], got %v", contents)
	}
}

// --- 6. ConvertDashScopeStreamToOpenAI 测试 ---

func TestConvertDashScopeStreamToOpenAI(t *testing.T) {
	mCtx := &mockHttpContext{
		requestId: "req-conv-stream-001",
	}
	ai_convert.SetAIModel(mCtx, "qwen-turbo")

	t.Run("basic stream chunk", func(t *testing.T) {
		resp := &dashScopeResponse{
			Output: dashScopeOutput{
				Choices: []dashScopeChoice{
					{
						FinishReason: "",
						Message: dashScopeMessage{
							Role:    "assistant",
							Content: "Hello stream",
						},
					},
				},
			},
			RequestID: "req-conv-1",
		}
		chunks := convertDashScopeStreamToOpenAI(mCtx, resp)
		if len(chunks) != 1 {
			t.Fatalf("expected 1 chunk, got %d", len(chunks))
		}
		if chunks[0].Object != "chat.completion.chunk" {
			t.Errorf("expected object chat.completion.chunk, got %s", chunks[0].Object)
		}
		if chunks[0].Choices[0].Delta.Content != "Hello stream" {
			t.Errorf("expected content 'Hello stream', got %s", chunks[0].Choices[0].Delta.Content)
		}
		if chunks[0].Choices[0].Delta.Role != "assistant" {
			t.Errorf("expected role assistant, got %s", chunks[0].Choices[0].Delta.Role)
		}
	})

	t.Run("chunk with usage", func(t *testing.T) {
		resp := &dashScopeResponse{
			Output: dashScopeOutput{
				Choices: []dashScopeChoice{
					{
						FinishReason: "stop",
						Message: dashScopeMessage{
							Role:    "assistant",
							Content: "final",
						},
					},
				},
			},
			Usage: dashScopeUsage{
				InputTokens:  15,
				OutputTokens: 25,
				TotalTokens:  40,
			},
		}
		chunks := convertDashScopeStreamToOpenAI(mCtx, resp)
		// 应当有 2 个 chunk: 内容 chunk + usage chunk
		if len(chunks) != 2 {
			t.Fatalf("expected 2 chunks (content + usage), got %d", len(chunks))
		}
		// 第二个 chunk 是 usage chunk
		if chunks[1].Usage == nil {
			t.Fatalf("expected second chunk to have usage")
		}
		if chunks[1].Usage.PromptTokens != 15 {
			t.Errorf("expected prompt tokens 15, got %d", chunks[1].Usage.PromptTokens)
		}
		if chunks[1].Usage.CompletionTokens != 25 {
			t.Errorf("expected completion tokens 25, got %d", chunks[1].Usage.CompletionTokens)
		}
		if len(chunks[1].Choices) != 0 {
			t.Errorf("expected usage chunk to have 0 choices, got %d", len(chunks[1].Choices))
		}
		// 验证上下文 token 标签
		if ai_convert.GetAIModelInputToken(mCtx) != 15 {
			t.Errorf("expected input token 15, got %d", ai_convert.GetAIModelInputToken(mCtx))
		}
		if ai_convert.GetAIModelTotalToken(mCtx) != 40 {
			t.Errorf("expected total token 40, got %d", ai_convert.GetAIModelTotalToken(mCtx))
		}
	})

	t.Run("empty role defaults to assistant", func(t *testing.T) {
		resp := &dashScopeResponse{
			Output: dashScopeOutput{
				Choices: []dashScopeChoice{
					{
						Message: dashScopeMessage{
							Role:    "",
							Content: "test",
						},
					},
				},
			},
		}
		chunks := convertDashScopeStreamToOpenAI(mCtx, resp)
		if chunks[0].Choices[0].Delta.Role != "assistant" {
			t.Errorf("expected default role assistant, got %s", chunks[0].Choices[0].Delta.Role)
		}
	})
}

// --- 7. ConvertDashScopeToOpenAI 测试 ---

func TestConvertDashScopeToOpenAI(t *testing.T) {
	mCtx := &mockHttpContext{
		requestId: "req-conv-non-stream-001",
	}
	ai_convert.SetAIModel(mCtx, "qwen-max")

	t.Run("choices with tool calls", func(t *testing.T) {
		resp := dashScopeResponse{
			Output: dashScopeOutput{
				Choices: []dashScopeChoice{
					{
						FinishReason: "tool_calls",
						Message: dashScopeMessage{
							Role: "assistant",
							ToolCalls: []openai.ToolCall{
								{
									ID:   "call_1",
									Type: openai.ToolTypeFunction,
									Function: openai.FunctionCall{
										Name:      "get_weather",
										Arguments: `{"city":"beijing"}`,
									},
								},
							},
						},
					},
				},
			},
			RequestID: "req-tool-1",
		}
		body, _ := json.Marshal(resp)
		out, err := convertDashScopeToOpenAI(mCtx, body)
		if err != nil {
			t.Fatalf("convertDashScopeToOpenAI failed: %v", err)
		}

		var openaiResp openai.ChatCompletionResponse
		if err := json.Unmarshal(out, &openaiResp); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if len(openaiResp.Choices) != 1 {
			t.Fatalf("expected 1 choice, got %d", len(openaiResp.Choices))
		}
		if openaiResp.Choices[0].FinishReason != openai.FinishReasonToolCalls {
			t.Errorf("expected finish reason tool_calls, got %v", openaiResp.Choices[0].FinishReason)
		}
		if len(openaiResp.Choices[0].Message.ToolCalls) != 1 {
			t.Fatalf("expected 1 tool call, got %d", len(openaiResp.Choices[0].Message.ToolCalls))
		}
		if openaiResp.Choices[0].Message.ToolCalls[0].Function.Name != "get_weather" {
			t.Errorf("expected function name get_weather, got %s", openaiResp.Choices[0].Message.ToolCalls[0].Function.Name)
		}
	})

	t.Run("request_id fallback to context", func(t *testing.T) {
		resp := dashScopeResponse{
			Output: dashScopeOutput{
				Text:         "fallback test",
				FinishReason: "stop",
			},
			RequestID: "",
		}
		body, _ := json.Marshal(resp)
		out, _ := convertDashScopeToOpenAI(mCtx, body)

		var openaiResp openai.ChatCompletionResponse
		_ = json.Unmarshal(out, &openaiResp)
		expectedID := "chatcmpl-" + mCtx.requestId
		if openaiResp.ID != expectedID {
			t.Errorf("expected ID %s, got %s", expectedID, openaiResp.ID)
		}
	})

	t.Run("multimodal content", func(t *testing.T) {
		resp := dashScopeResponse{
			Output: dashScopeOutput{
				Choices: []dashScopeChoice{
					{
						FinishReason: "stop",
						Message: dashScopeMessage{
							Role: "assistant",
							Content: []interface{}{
								map[string]interface{}{"text": "part1"},
								map[string]interface{}{"text": "part2"},
								map[string]interface{}{"image": "url"},
							},
						},
					},
				},
			},
		}
		body, _ := json.Marshal(resp)
		out, _ := convertDashScopeToOpenAI(mCtx, body)

		var openaiResp openai.ChatCompletionResponse
		_ = json.Unmarshal(out, &openaiResp)
		if openaiResp.Choices[0].Message.Content != "part1part2" {
			t.Errorf("expected content 'part1part2', got %s", openaiResp.Choices[0].Message.Content)
		}
	})

	t.Run("empty role defaults to assistant", func(t *testing.T) {
		resp := dashScopeResponse{
			Output: dashScopeOutput{
				Choices: []dashScopeChoice{
					{
						FinishReason: "stop",
						Message: dashScopeMessage{
							Role:    "",
							Content: "no role",
						},
					},
				},
			},
		}
		body, _ := json.Marshal(resp)
		out, _ := convertDashScopeToOpenAI(mCtx, body)

		var openaiResp openai.ChatCompletionResponse
		_ = json.Unmarshal(out, &openaiResp)
		if openaiResp.Choices[0].Message.Role != "assistant" {
			t.Errorf("expected role assistant, got %s", openaiResp.Choices[0].Message.Role)
		}
	})
}
