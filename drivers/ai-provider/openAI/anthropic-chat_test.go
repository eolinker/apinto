package openAI

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	ai_convert "github.com/eolinker/apinto/ai-convert"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	openai "github.com/sashabaranov/go-openai"
)

func TestNewAnthropicChat(t *testing.T) {
	t.Run("default base url", func(t *testing.T) {
		driver, err := NewAnthropicChat(provider, "test-api-key", "", ai_convert.ModelTypeAnthropicChat, 5*time.Second)
		if err != nil {
			t.Fatalf("failed to create AnthropicChat: %v", err)
		}
		chat, ok := driver.(*AnthropicChat)
		if !ok {
			t.Fatalf("driver is not *AnthropicChat")
		}
		if chat.apikey != "test-api-key" {
			t.Errorf("expected apikey test-api-key, got %s", chat.apikey)
		}
		if chat.Provider() != provider {
			t.Errorf("expected provider %s, got %s", provider, chat.Provider())
		}
		if chat.ModelType() != ai_convert.ModelTypeAnthropicChat {
			t.Errorf("expected modelType %v, got %v", ai_convert.ModelTypeAnthropicChat, chat.ModelType())
		}
		if chat.path != "/v1/chat/completions" {
			t.Errorf("expected path /v1/chat/completions, got %s", chat.path)
		}
	})

	t.Run("custom base url with prefix", func(t *testing.T) {
		driver, err := NewAnthropicChat(provider, "test-api-key", "https://api.custom.com/custom/v1", ai_convert.ModelTypeAnthropicChat, 5*time.Second)
		if err != nil {
			t.Fatalf("failed to create AnthropicChat: %v", err)
		}
		chat := driver.(*AnthropicChat)
		if chat.path != "/custom/v1/chat/completions" {
			t.Errorf("expected path /custom/v1/chat/completions, got %s", chat.path)
		}
	})
}

func TestConvertAnthropicToOpenAIRequest(t *testing.T) {
	temp := float32(0.7)
	topP := float32(0.9)
	req := &AnthropicMessageRequest{
		Model:         "gpt-4o",
		MaxTokens:     1000,
		Stream:        true,
		Temperature:   &temp,
		TopP:          &topP,
		StopSequences: []string{"STOP"},
		System:        "You are a helpful assistant.",
		Messages: []AnthropicMessage{
			{
				Role:    "user",
				Content: "Hello OpenAI",
			},
			{
				Role: "assistant",
				Content: []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": "Let me call a tool.",
					},
					map[string]interface{}{
						"type": "tool_use",
						"id":   "call_123",
						"name": "get_weather",
						"input": map[string]interface{}{
							"city": "Beijing",
						},
					},
				},
			},
			{
				Role: "user",
				Content: []interface{}{
					map[string]interface{}{
						"type":        "tool_result",
						"tool_use_id": "call_123",
						"content":     `{"temp":"25C"}`,
					},
				},
			},
		},
		Tools: []AnthropicTool{
			{
				Name:        "get_weather",
				Description: "Get city weather",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"city": map[string]interface{}{"type": "string"},
					},
				},
			},
		},
		ToolChoice: "auto",
	}

	openaiReq := convertAnthropicToOpenAIRequest(req, nil)
	if openaiReq == nil {
		t.Fatalf("expected non-nil openaiReq")
	}
	if openaiReq.Model != "gpt-4o" {
		t.Errorf("expected model gpt-4o, got %s", openaiReq.Model)
	}
	if openaiReq.MaxTokens != 1000 {
		t.Errorf("expected max_tokens 1000, got %d", openaiReq.MaxTokens)
	}
	if !openaiReq.Stream || openaiReq.StreamOptions == nil || !openaiReq.StreamOptions.IncludeUsage {
		t.Errorf("expected stream with include_usage")
	}
	if len(openaiReq.Messages) != 4 {
		t.Fatalf("expected 4 messages (1 system + 3 converted), got %d", len(openaiReq.Messages))
	}
	if openaiReq.Messages[0].Role != openai.ChatMessageRoleSystem || openaiReq.Messages[0].Content != "You are a helpful assistant." {
		t.Errorf("system message mismatch: %+v", openaiReq.Messages[0])
	}
	if openaiReq.Messages[1].Role != "user" || openaiReq.Messages[1].Content != "Hello OpenAI" {
		t.Errorf("user message mismatch: %+v", openaiReq.Messages[1])
	}
	if openaiReq.Messages[2].Role != "assistant" || len(openaiReq.Messages[2].ToolCalls) != 1 {
		t.Errorf("assistant message mismatch: %+v", openaiReq.Messages[2])
	}
	if openaiReq.Messages[3].Role != "tool" || openaiReq.Messages[3].ToolCallID != "call_123" {
		t.Errorf("tool_result message mismatch: %+v", openaiReq.Messages[3])
	}
	if len(openaiReq.Tools) != 1 || openaiReq.Tools[0].Function.Name != "get_weather" {
		t.Errorf("tools mismatch: %+v", openaiReq.Tools)
	}
}

func TestConvertOpenAIToAnthropicResponse(t *testing.T) {
	openaiResp := &openai.ChatCompletionResponse{
		ID:    "chatcmpl-test-id",
		Model: "gpt-4o",
		Choices: []openai.ChatCompletionChoice{
			{
				Index: 0,
				Message: openai.ChatCompletionMessage{
					Role:    "assistant",
					Content: "Here is your answer.",
					ToolCalls: []openai.ToolCall{
						{
							ID:   "call_abc",
							Type: openai.ToolTypeFunction,
							Function: openai.FunctionCall{
								Name:      "calc",
								Arguments: `{"a":1,"b":2}`,
							},
						},
					},
				},
				FinishReason: openai.FinishReasonToolCalls,
			},
		},
		Usage: openai.Usage{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
	}

	resp := convertOpenAIToAnthropicResponse(openaiResp, "default-model")
	if resp == nil {
		t.Fatalf("expected non-nil response")
	}
	if resp.ID != "chatcmpl-test-id" {
		t.Errorf("expected ID chatcmpl-test-id, got %s", resp.ID)
	}
	if resp.Model != "gpt-4o" {
		t.Errorf("expected model gpt-4o, got %s", resp.Model)
	}
	if resp.Usage.InputTokens != 10 || resp.Usage.OutputTokens != 20 {
		t.Errorf("usage mismatch: %+v", resp.Usage)
	}
	if len(resp.Content) != 2 {
		t.Fatalf("expected 2 content blocks, got %d", len(resp.Content))
	}
	if resp.Content[0].Type != "text" || resp.Content[0].Text != "Here is your answer." {
		t.Errorf("text content mismatch: %+v", resp.Content[0])
	}
	if resp.Content[1].Type != "tool_use" || resp.Content[1].Name != "calc" || resp.Content[1].ID != "call_abc" {
		t.Errorf("tool_use content mismatch: %+v", resp.Content[1])
	}
	if resp.StopReason == nil || *resp.StopReason != "tool_use" {
		t.Errorf("expected stop_reason tool_use, got %v", resp.StopReason)
	}
}

func TestAnthropicChat_RequestConvert(t *testing.T) {
	driver, _ := NewAnthropicChat(provider, "test-api-key", "", ai_convert.ModelTypeAnthropicChat, 5*time.Second)
	chat := driver.(*AnthropicChat)

	req := AnthropicMessageRequest{
		Model: "gpt-4o",
		Messages: []AnthropicMessage{
			{Role: "user", Content: "Hello"},
		},
		MaxTokens: 500,
	}
	body, _ := json.Marshal(req)

	mCtx := &mockHttpContext{
		proxy: &mockRequest{
			header: &mockHeader{},
			body:   &mockBody{body: body},
			uri:    &mockURI{},
		},
		requestId: "req-123",
	}

	err := chat.RequestConvert(mCtx, map[string]interface{}{})
	if err != nil {
		t.Fatalf("RequestConvert failed: %v", err)
	}

	if mCtx.proxy.header.Get("Authorization") != "Bearer test-api-key" {
		t.Errorf("expected auth header Bearer test-api-key, got %s", mCtx.proxy.header.Get("Authorization"))
	}
	if mCtx.proxy.uri.path != "/v1/chat/completions" {
		t.Errorf("expected path /v1/chat/completions, got %s", mCtx.proxy.uri.path)
	}

	var convertedReq openai.ChatCompletionRequest
	if err := json.Unmarshal(mCtx.proxy.body.body, &convertedReq); err != nil {
		t.Fatalf("failed to unmarshal converted request: %v", err)
	}
	if convertedReq.Model != "gpt-4o" || convertedReq.MaxTokens != 500 {
		t.Errorf("converted request mismatch: %+v", convertedReq)
	}
}

func TestAnthropicChat_ResponseConvert(t *testing.T) {
	driver, _ := NewAnthropicChat(provider, "test-api-key", "", ai_convert.ModelTypeAnthropicChat, 5*time.Second)
	chat := driver.(*AnthropicChat)

	openaiResp := openai.ChatCompletionResponse{
		ID:    "chatcmpl-test",
		Model: "gpt-4o",
		Choices: []openai.ChatCompletionChoice{
			{
				Index: 0,
				Message: openai.ChatCompletionMessage{
					Role:    "assistant",
					Content: "Hello human",
				},
				FinishReason: openai.FinishReasonStop,
			},
		},
		Usage: openai.Usage{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
	}
	body, _ := json.Marshal(openaiResp)

	mCtx := &mockHttpContext{
		response: &mockResponse{
			statusCode: 200,
			body:       body,
		},
		requestId: "req-456",
	}

	err := chat.ResponseConvert(mCtx)
	if err != nil {
		t.Fatalf("ResponseConvert failed: %v", err)
	}

	var anthropicResp AnthropicResponse
	if err := json.Unmarshal(mCtx.response.body, &anthropicResp); err != nil {
		t.Fatalf("failed to unmarshal anthropic response: %v", err)
	}
	if anthropicResp.ID != "chatcmpl-test" || anthropicResp.Model != "gpt-4o" {
		t.Errorf("response mismatch: %+v", anthropicResp)
	}
	if len(anthropicResp.Content) != 1 || anthropicResp.Content[0].Text != "Hello human" {
		t.Errorf("content mismatch: %+v", anthropicResp.Content)
	}
}

func TestAnthropicChat_StreamHandler(t *testing.T) {
	driver, _ := NewAnthropicChat(provider, "test-key", "", ai_convert.ModelTypeAnthropicChat, 5*time.Second)
	chat := driver.(*AnthropicChat)

	mockCtx := &mockHttpContext{
		labels:    make(map[string]string),
		requestId: "stream-req-123",
	}
	ai_convert.SetAIModel(mockCtx, "gpt-4o")

	chunk1 := `data: {"id":"chatcmpl-stream","model":"gpt-4o","choices":[{"index":0,"delta":{"role":"assistant","content":"Hello"},"finish_reason":null}]}` + "\n\n"
	out1, err := chat.streamHandler(mockCtx, []byte(chunk1))
	if err != nil {
		t.Fatalf("streamHandler error: %v", err)
	}
	out1Str := string(out1)
	if !strings.Contains(out1Str, "event: message_start") {
		t.Errorf("expected message_start in output: %s", out1Str)
	}
	if !strings.Contains(out1Str, "event: content_block_start") {
		t.Errorf("expected content_block_start in output: %s", out1Str)
	}
	if !strings.Contains(out1Str, "event: content_block_delta") {
		t.Errorf("expected content_block_delta in output: %s", out1Str)
	}

	chunkDone := `data: [DONE]` + "\n\n"
	outDone, err := chat.streamHandler(mockCtx, []byte(chunkDone))
	if err != nil {
		t.Fatalf("streamHandler error on DONE: %v", err)
	}
	outDoneStr := string(outDone)
	if !strings.Contains(outDoneStr, "event: content_block_stop") {
		t.Errorf("expected content_block_stop in output: %s", outDoneStr)
	}
	if !strings.Contains(outDoneStr, "event: message_delta") {
		t.Errorf("expected message_delta in output: %s", outDoneStr)
	}
	if !strings.Contains(outDoneStr, "event: message_stop") {
		t.Errorf("expected message_stop in output: %s", outDoneStr)
	}
}

// --- Mock Helpers ---

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
	return nil
}
