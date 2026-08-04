package google

import (
	"bytes"
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

func TestCanonicalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "string",
			input:    "hello",
			expected: `"hello"`,
		},
		{
			name:     "number",
			input:    123.45,
			expected: `123.45`,
		},
		{
			name:     "slice",
			input:    []interface{}{"b", "a", 1},
			expected: `["b","a",1]`,
		},
		{
			name: "map sorted",
			input: map[string]interface{}{
				"z": 1,
				"a": "hello",
				"m": []interface{}{2, 1},
			},
			expected: `{"a":"hello","m":[2,1],"z":1}`,
		},
		{
			name: "nested map",
			input: map[string]interface{}{
				"nested": map[string]interface{}{
					"y": "y",
					"x": "x",
				},
			},
			expected: `{"nested":{"x":"x","y":"y"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := canonicalJSON(tt.input)
			if actual != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, actual)
			}
		})
	}
}

func TestThoughtSignatureKey(t *testing.T) {
	name := "my_func"
	args := map[string]interface{}{
		"b": 2,
		"a": 1,
	}
	key1 := thoughtSignatureKey(name, args)

	args2 := map[string]interface{}{
		"a": 1,
		"b": 2,
	}
	key2 := thoughtSignatureKey(name, args2)

	if key1 != key2 {
		t.Errorf("expected same key for identical but unsorted maps, got %s and %s", key1, key2)
	}

	if key1 == "" {
		t.Errorf("key should not be empty")
	}
}

func TestThoughtSignatureCache(t *testing.T) {
	name := "cached_func"
	args := map[string]interface{}{"p": "p"}
	sig := "test-signature-123"

	cacheThoughtSignature(name, args, sig)

	retrieved := lookupCachedThoughtSignature(name, args)
	if retrieved != sig {
		t.Errorf("expected signature %s, got %s", sig, retrieved)
	}

	emptySig := lookupCachedThoughtSignature("other_func", args)
	if emptySig != "" {
		t.Errorf("expected empty string, got %s", emptySig)
	}
}

func TestConvertSchemaToGemini(t *testing.T) {
	schema := map[string]interface{}{
		"$schema":              "https://json-schema.org/draft/2020-12/schema",
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type":             "string",
				"exclusiveMinimum": 1,
			},
			"items": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id": map[string]interface{}{
							"type": "integer",
						},
					},
				},
			},
		},
	}

	convertSchemaToGemini(schema)

	// 验证 Gemini 不支持的 schema 字段是否被移除
	if _, ok := schema["$schema"]; ok {
		t.Errorf("$schema should be removed")
	}
	if _, ok := schema["additionalProperties"]; ok {
		t.Errorf("additionalProperties should be removed")
	}

	// 验证 type 变为了大写
	if schema["type"] != "OBJECT" {
		t.Errorf("expected type to be OBJECT, got %v", schema["type"])
	}

	props := schema["properties"].(map[string]interface{})
	nameProp := props["name"].(map[string]interface{})
	if nameProp["type"] != "STRING" {
		t.Errorf("expected property type to be STRING, got %v", nameProp["type"])
	}
	if _, ok := nameProp["exclusiveMinimum"]; ok {
		t.Errorf("exclusiveMinimum should be removed")
	}

	itemsProp := props["items"].(map[string]interface{})
	if itemsProp["type"] != "ARRAY" {
		t.Errorf("expected items property type to be ARRAY, got %v", itemsProp["type"])
	}

	subItems := itemsProp["items"].(map[string]interface{})
	if subItems["type"] != "OBJECT" {
		t.Errorf("expected sub-items type to be OBJECT, got %v", subItems["type"])
	}
}

// TestConvertSchemaToGeminiAnyOf 验证 anyOf / allOf / oneOf 组合关键字内
// 嵌套的不被 Gemini 支持的字段（additionalProperties / propertyNames / const 等）
// 也会被递归移除。
func TestConvertSchemaToGeminiAnyOf(t *testing.T) {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"flex": map[string]interface{}{
				"anyOf": []interface{}{
					map[string]interface{}{
						"type":                 "object",
						"additionalProperties": false,
						"properties": map[string]interface{}{
							"k": map[string]interface{}{
								"type": "string",
								"const": "fixed",
							},
						},
					},
					map[string]interface{}{
						"type":          "object",
						"propertyNames": map[string]interface{}{"pattern": "^[a-z]+$"},
					},
				},
			},
		},
	}

	convertSchemaToGemini(schema)

	flex := schema["properties"].(map[string]interface{})["flex"].(map[string]interface{})
	anyOf := flex["anyOf"].([]interface{})
	first := anyOf[0].(map[string]interface{})
	if _, ok := first["additionalProperties"]; ok {
		t.Errorf("additionalProperties inside anyOf[0] should be removed")
	}
	kProp := first["properties"].(map[string]interface{})["k"].(map[string]interface{})
	if _, ok := kProp["const"]; ok {
		t.Errorf("const inside anyOf[0].properties.k should be removed")
	}
	second := anyOf[1].(map[string]interface{})
	if _, ok := second["propertyNames"]; ok {
		t.Errorf("propertyNames inside anyOf[1] should be removed")
	}
	if first["type"] != "OBJECT" || second["type"] != "OBJECT" {
		t.Errorf("types inside anyOf should be uppercased")
	}
}

func TestExtractExtraContentSignatures(t *testing.T) {
	body := []byte(`{
		"messages": [
			{
				"role": "assistant",
				"tool_calls": [
					{
						"id": "tc-1",
						"extra_content": {
							"google": {
								"thought_signature": "sig-1"
							}
						}
					},
					{
						"id": "tc-2"
					}
				]
			},
			{
				"role": "user"
			}
		]
	}`)

	sigs := extractExtraContentSignatures(body)
	if sigs == nil {
		t.Fatalf("failed to extract signatures")
	}

	sig := lookupExtraSignature(sigs, 0, 0)
	if sig != "sig-1" {
		t.Errorf("expected sig-1, got %s", sig)
	}

	sigEmpty := lookupExtraSignature(sigs, 0, 1)
	if sigEmpty != "" {
		t.Errorf("expected empty signature, got %s", sigEmpty)
	}

	sigNotExist := lookupExtraSignature(sigs, 1, 0)
	if sigNotExist != "" {
		t.Errorf("expected empty signature for msg 1, got %s", sigNotExist)
	}

	// 测试无效 JSON
	invalidSigs := extractExtraContentSignatures([]byte("{invalid"))
	if invalidSigs != nil {
		t.Errorf("expected nil for invalid JSON, got %v", invalidSigs)
	}
}

func TestMimeTypeFromURL(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"https://example.com/image.png", "image/png"},
		{"https://example.com/image.PNG", "image/png"},
		{"https://example.com/image.jpg?width=100#anchor", "image/jpeg"},
		{"https://example.com/image.jpeg", "image/jpeg"},
		{"https://example.com/image.gif", "image/gif"},
		{"https://example.com/image.webp", "image/webp"},
		{"https://example.com/image.unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			actual := mimeTypeFromURL(tt.url)
			if actual != tt.expected {
				t.Errorf("expected %s, got %s for %s", tt.expected, actual, tt.url)
			}
		})
	}
}

func TestInjectToolCallSignatures(t *testing.T) {
	body := []byte(`{
		"choices": [
			{
				"index": 0,
				"message": {
					"tool_calls": [
						{"id": "call-1", "type": "function"}
					]
				}
			}
		]
	}`)

	signatures := map[string]string{
		"call-1": "signature-123",
	}

	injected := injectToolCallSignatures(body, signatures)

	var result map[string]interface{}
	err := json.Unmarshal(injected, &result)
	if err != nil {
		t.Fatalf("failed to unmarshal injected body: %v", err)
	}

	choices := result["choices"].([]interface{})
	choice := choices[0].(map[string]interface{})
	message := choice["message"].(map[string]interface{})
	toolCalls := message["tool_calls"].([]interface{})
	tc := toolCalls[0].(map[string]interface{})

	extra, ok := tc["extra_content"].(map[string]interface{})
	if !ok {
		t.Fatalf("extra_content not found")
	}
	google := extra["google"].(map[string]interface{})
	sig := google["thought_signature"].(string)

	if sig != "signature-123" {
		t.Errorf("expected signature-123, got %s", sig)
	}

	// 测试空的 signatures 应当原样返回
	original := injectToolCallSignatures(body, nil)
	if !bytes.Equal(original, body) {
		t.Errorf("expected body to be unchanged when signatures are empty")
	}

	// 测试无效 JSON 应当原样返回
	invalidBody := []byte("{invalid")
	injectedInvalid := injectToolCallSignatures(invalidBody, signatures)
	if !bytes.Equal(injectedInvalid, invalidBody) {
		t.Errorf("expected invalid body to be returned unchanged")
	}
}

// --- 2. NewOpenAIChat 构造测试 ---

func TestNewOpenAIChat(t *testing.T) {
	// 默认 BaseUrl 为空
	driver, err := NewOpenAIChat("google", "my-key", "", ai_convert.ModelTypeOpenAIChat, 5*time.Second)
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
	if chat.provider != "google" {
		t.Errorf("expected provider google, got %s", chat.provider)
	}
	if chat.modelType != ai_convert.ModelTypeOpenAIChat {
		t.Errorf("expected modelType %v, got %v", ai_convert.ModelTypeOpenAIChat, chat.modelType)
	}
	if !strings.HasSuffix(chat.path, "/models") {
		t.Errorf("expected default path to end with /models, got %s", chat.path)
	}

	// 指定 BaseUrl 带有子路径
	driverWithUrl, err := NewOpenAIChat("google", "my-key", "https://proxy.example.com/custom-prefix", ai_convert.ModelTypeOpenAIChat, 5*time.Second)
	if err != nil {
		t.Fatalf("failed to create driver: %v", err)
	}
	chatWithUrl := driverWithUrl.(*OpenAIChat)
	if !strings.Contains(chatWithUrl.path, "/custom-prefix") {
		t.Errorf("expected path to contain custom-prefix, got %s", chatWithUrl.path)
	}

	// 无效 BaseUrl
	_, err = NewOpenAIChat("google", "my-key", "::invalid-url::", ai_convert.ModelTypeOpenAIChat, 5*time.Second)
	if err == nil {
		t.Errorf("expected error for invalid URL, got nil")
	}
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
	chat, err := NewOpenAIChat("google", "api-key-xyz", "https://api.generativeai.com", ai_convert.ModelTypeOpenAIChat, 5*time.Second)
	if err != nil {
		t.Fatalf("failed to create driver: %v", err)
	}

	// 构造 Mock 请求
	chatReq := openai.ChatCompletionRequest{
		Model: "gemini-1.5-pro",
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "system",
				Content: "You are an AI assistant.",
			},
			{
				Role:    "user",
				Content: "Hello!",
			},
		},
		Temperature: 0.7,
		MaxTokens:   150,
	}

	reqBytes, _ := json.Marshal(chatReq)

	mCtx := &mockHttpContext{
		proxy: &mockRequest{
			header: &mockHeader{},
			body: &mockBody{
				body: reqBytes,
			},
			uri: &mockURI{},
		},
		requestId: "req-123456",
	}

	// 设置 AI 模型标签，GetAIModel 会用到
	ai_convert.SetAIModel(mCtx, "gemini-1.5-pro")

	extender := map[string]interface{}{}
	err = chat.RequestConvert(mCtx, extender)
	if err != nil {
		t.Fatalf("RequestConvert failed: %v", err)
	}

	// 验证请求头是否正确设置
	apiKeyHeader := mCtx.proxy.header.Get("x-goog-api-key")
	if apiKeyHeader != "api-key-xyz" {
		t.Errorf("expected API key header api-key-xyz, got %s", apiKeyHeader)
	}

	// 验证代理的 Body 是否被转换为了 Gemini 格式
	var geminiReq GeminiRequest
	err = json.Unmarshal(mCtx.proxy.body.body, &geminiReq)
	if err != nil {
		t.Fatalf("failed to unmarshal converted body to GeminiRequest: %v", err)
	}

	if geminiReq.SystemInstruction == nil || len(geminiReq.SystemInstruction.Parts) != 1 || geminiReq.SystemInstruction.Parts[0].Text != "You are an AI assistant." {
		t.Errorf("system instruction mismatch: %v", geminiReq.SystemInstruction)
	}

	if len(geminiReq.Contents) != 1 || geminiReq.Contents[0].Role != "user" || geminiReq.Contents[0].Parts[0].Text != "Hello!" {
		t.Errorf("contents mismatch: %v", geminiReq.Contents)
	}

	if geminiReq.GenerationConfig == nil || *geminiReq.GenerationConfig.Temperature != 0.7 || *geminiReq.GenerationConfig.MaxOutputTokens != 150 {
		t.Errorf("generation config mismatch")
	}

	// 验证路径设置
	expectedPath := "/v1beta/models/gemini-1.5-pro:generateContent"
	if mCtx.proxy.uri.path != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, mCtx.proxy.uri.path)
	}
}

func TestResponseConvert(t *testing.T) {
	chat, _ := NewOpenAIChat("google", "api-key-xyz", "", ai_convert.ModelTypeOpenAIChat, 5*time.Second)

	// 构造 Mock Gemini 响应
	geminiResp := GeminiResponse{
		Candidates: []GeminiCandidate{
			{
				Index: 0,
				Content: GeminiContent{
					Role: "model",
					Parts: []GeminiPart{
						{Text: "This is a response from Gemini."},
					},
				},
				FinishReason: "STOP",
			},
		},
		UsageMetadata: &GeminiUsageMetadata{
			PromptTokenCount:     10,
			CandidatesTokenCount: 20,
			TotalTokenCount:      30,
		},
	}
	respBytes, _ := json.Marshal(geminiResp)

	mCtx := &mockHttpContext{
		response: &mockResponse{
			statusCode: 200,
			body:       respBytes,
			headers:    map[string]string{"content-encoding": "utf-8"},
		},
		requestId: "req-123",
	}

	ai_convert.SetAIModel(mCtx, "gemini-1.5-pro")

	err := chat.ResponseConvert(mCtx)
	if err != nil {
		t.Fatalf("ResponseConvert failed: %v", err)
	}

	// 验证 Response 已经转为了 OpenAI 格式
	var openaiResp openai.ChatCompletionResponse
	err = json.Unmarshal(mCtx.response.body, &openaiResp)
	if err != nil {
		t.Fatalf("failed to unmarshal converted body to ChatCompletionResponse: %v", err)
	}

	if len(openaiResp.Choices) != 1 || openaiResp.Choices[0].Message.Content != "This is a response from Gemini." {
		t.Errorf("choices mismatch: %v", openaiResp.Choices)
	}

	if openaiResp.Choices[0].FinishReason != openai.FinishReasonStop {
		t.Errorf("expected finish reason to be stop, got %v", openaiResp.Choices[0].FinishReason)
	}

	if openaiResp.Usage.PromptTokens != 10 || openaiResp.Usage.CompletionTokens != 20 || openaiResp.Usage.TotalTokens != 30 {
		t.Errorf("usage mismatch")
	}

	// 验证上下文中的 Token 标签也被正确设置了
	if ai_convert.GetAIModelInputToken(mCtx) != 10 {
		t.Errorf("expected input token 10, got %d", ai_convert.GetAIModelInputToken(mCtx))
	}
	if ai_convert.GetAIModelOutputToken(mCtx) != 20 {
		t.Errorf("expected output token 20, got %d", ai_convert.GetAIModelOutputToken(mCtx))
	}
	if ai_convert.GetAIModelTotalToken(mCtx) != 30 {
		t.Errorf("expected total token 30, got %d", ai_convert.GetAIModelTotalToken(mCtx))
	}
}

func TestConvertErrorResponse(t *testing.T) {
	chat, _ := NewOpenAIChat("google", "api-key", "", ai_convert.ModelTypeOpenAIChat, 5*time.Second)

	geminiErr := `{
		"error": {
			"code": 400,
			"message": "API key not valid",
			"status": "INVALID_ARGUMENT"
		}
	}`

	mCtx := &mockHttpContext{
		response: &mockResponse{
			statusCode: 400,
			body:       []byte(geminiErr),
		},
	}

	// 私有方法，通过 type assertion 获取实体并测试
	chatImpl := chat.(*OpenAIChat)
	chatImpl.convertErrorResponse(mCtx)

	var openaiErr struct {
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Code    int    `json:"code"`
		} `json:"error"`
	}

	err := json.Unmarshal(mCtx.response.body, &openaiErr)
	if err != nil {
		t.Fatalf("failed to unmarshal converted error body: %v", err)
	}

	if openaiErr.Error.Message != "API key not valid" || openaiErr.Error.Type != "INVALID_ARGUMENT" || openaiErr.Error.Code != 400 {
		t.Errorf("error converted format mismatch: %v", openaiErr.Error)
	}
}

// --- 5. Stream Handler 集成测试 ---

func TestStreamHandler(t *testing.T) {
	chat, _ := NewOpenAIChat("google", "api-key", "", ai_convert.ModelTypeOpenAIChat, 5*time.Second)
	chatImpl := chat.(*OpenAIChat)

	mCtx := &mockHttpContext{
		requestId: "req-stream-456",
		labels:    map[string]string{},
	}
	ai_convert.SetAIModel(mCtx, "gemini-1.5-pro")

	// 构造分片的 SSE 数据。我们通过两次输入流，分别传入一部分来测试 remain 缓存和拼接
	chunk1 := []byte("data: {\"candidates\":[{\"index\":0,\"content\":{\"parts\":[{\"text\":\"Hello \"}]},\"finishReason\":\"\"}]}\n")
	chunk2 := []byte("data: {\"candidates\":[{\"index\":0,\"content\":{\"parts\":[{\"text\":\"World\"}]},\"finishReason\":\"STOP\"}],\"usageMetadata\":{\"promptTokenCount\":5,\"candidatesTokenCount\":5,\"totalTokenCount\":10}}\n\n")

	// 测试第一段（不完整，故意去掉换行或直接发 chunk1）
	out1, err := chatImpl.streamHandler(mCtx, chunk1)
	if err != nil {
		t.Fatalf("streamHandler failed: %v", err)
	}

	if len(out1) == 0 {
		t.Errorf("expected out1 output, got empty")
	}

	// 校验 out1 的数据格式是否为 SSE "data: ..."
	if !strings.HasPrefix(string(out1), "data:") {
		t.Errorf("expected out1 to start with 'data:', got %s", string(out1))
	}

	var streamResp1 openai.ChatCompletionStreamResponse
	cleanData1 := strings.TrimPrefix(strings.TrimSuffix(string(out1), "\n\n"), "data: ")
	err = json.Unmarshal([]byte(cleanData1), &streamResp1)
	if err != nil {
		t.Fatalf("failed to unmarshal stream output1: %v, data: %s", err, cleanData1)
	}
	if streamResp1.Choices[0].Delta.Content != "Hello " {
		t.Errorf("expected content 'Hello ', got %s", streamResp1.Choices[0].Delta.Content)
	}

	// 输入第二段（含 STOP 状态和 usage）
	out2, err := chatImpl.streamHandler(mCtx, chunk2)
	if err != nil {
		t.Fatalf("streamHandler chunk2 failed: %v", err)
	}

	if !strings.Contains(string(out2), "data: [DONE]") {
		t.Errorf("expected out2 to contain data: [DONE]")
	}

	// 验证 token 计数也被正确解析
	if ai_convert.GetAIModelInputToken(mCtx) != 5 {
		t.Errorf("expected input tokens = 5, got %d", ai_convert.GetAIModelInputToken(mCtx))
	}
}
