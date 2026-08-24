package google

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	ai_convert "github.com/eolinker/apinto/ai-convert"
)

func TestNewAnthropicChat(t *testing.T) {
	driver, err := NewAnthropicChat("google-generative-ai", "test-key", "", ai_convert.ModelTypeAnthropicChat, 5*time.Second)
	if err != nil {
		t.Fatalf("failed to create AnthropicChat: %v", err)
	}

	chat, ok := driver.(*AnthropicChat)
	if !ok {
		t.Fatalf("driver is not *AnthropicChat")
	}

	if chat.apikey != "test-key" {
		t.Errorf("expected apikey test-key, got %s", chat.apikey)
	}
	if chat.Provider() != "google-generative-ai" {
		t.Errorf("expected provider google-generative-ai, got %s", chat.Provider())
	}
	if chat.ModelType() != ai_convert.ModelTypeAnthropicChat {
		t.Errorf("expected modelType %v, got %v", ai_convert.ModelTypeAnthropicChat, chat.ModelType())
	}
	if !strings.HasSuffix(chat.path, "/models") {
		t.Errorf("expected default path to end with /models, got %s", chat.path)
	}
}

func TestConvertAnthropicToGeminiRequest(t *testing.T) {
	temp := float32(0.7)
	topP := float32(0.9)
	req := &AnthropicMessageRequest{
		Model:       "gemini-1.5-pro",
		MaxTokens:   1000,
		Temperature: &temp,
		TopP:        &topP,
		System:      "You are a helpful assistant.",
		Messages: []AnthropicMessage{
			{
				Role:    "user",
				Content: "Hello world",
			},
			{
				Role: "assistant",
				Content: []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": "I will check the weather.",
					},
					map[string]interface{}{
						"type": "tool_use",
						"id":   "call_123_ts_testsignature",
						"name": "get_weather",
						"input": map[string]interface{}{
							"location": "Paris",
						},
					},
				},
			},
			{
				Role: "user",
				Content: []interface{}{
					map[string]interface{}{
						"type":        "tool_result",
						"tool_use_id": "call_123_ts_testsignature",
						"content":     `{"temperature": "22C"}`,
					},
				},
			},
		},
		Tools: []AnthropicTool{
			{
				Name:        "get_weather",
				Description: "Get weather in a city",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"location": map[string]interface{}{
							"type": "string",
						},
					},
					"required": []interface{}{"location"},
				},
			},
		},
		ToolChoice: "auto",
	}

	geminiReq := convertAnthropicToGeminiRequest(req)
	if geminiReq == nil {
		t.Fatalf("expected non-nil GeminiRequest")
	}

	// Verify System
	if geminiReq.SystemInstruction == nil || len(geminiReq.SystemInstruction.Parts) == 0 {
		t.Fatalf("expected systemInstruction")
	}
	if geminiReq.SystemInstruction.Parts[0].Text != "You are a helpful assistant." {
		t.Errorf("system instruction mismatch: %s", geminiReq.SystemInstruction.Parts[0].Text)
	}

	// Verify Contents
	if len(geminiReq.Contents) != 3 {
		t.Fatalf("expected 3 contents, got %d", len(geminiReq.Contents))
	}
	if geminiReq.Contents[0].Role != "user" || geminiReq.Contents[0].Parts[0].Text != "Hello world" {
		t.Errorf("user message mismatch: %+v", geminiReq.Contents[0])
	}
	if geminiReq.Contents[1].Role != "model" || len(geminiReq.Contents[1].Parts) != 2 {
		t.Fatalf("model message mismatch: %+v", geminiReq.Contents[1])
	}
	if geminiReq.Contents[1].Parts[1].FunctionCall == nil || geminiReq.Contents[1].Parts[1].FunctionCall.Name != "get_weather" {
		t.Errorf("functionCall mismatch: %+v", geminiReq.Contents[1].Parts[1])
	}
	if geminiReq.Contents[1].Parts[1].ThoughtSignature != "testsignature" {
		t.Errorf("expected thoughtSignature testsignature, got %s", geminiReq.Contents[1].Parts[1].ThoughtSignature)
	}

	// Verify tool_result
	if geminiReq.Contents[2].Role != "user" || geminiReq.Contents[2].Parts[0].FunctionResponse == nil {
		t.Fatalf("tool_result mismatch: %+v", geminiReq.Contents[2])
	}
	if geminiReq.Contents[2].Parts[0].FunctionResponse.Name != "get_weather" {
		t.Errorf("functionResponse name mismatch: %s", geminiReq.Contents[2].Parts[0].FunctionResponse.Name)
	}

	// Verify Tools
	if len(geminiReq.Tools) == 0 || len(geminiReq.Tools[0].FunctionDeclarations) == 0 {
		t.Fatalf("expected tools declarations")
	}
	if geminiReq.Tools[0].FunctionDeclarations[0].Name != "get_weather" {
		t.Errorf("tool name mismatch")
	}

	// Verify ToolConfig
	if geminiReq.ToolConfig == nil || geminiReq.ToolConfig.FunctionCallingConfig == nil || geminiReq.ToolConfig.FunctionCallingConfig.Mode != "AUTO" {
		t.Errorf("toolConfig mismatch: %+v", geminiReq.ToolConfig)
	}

	// Verify GenerationConfig
	if geminiReq.GenerationConfig == nil || *geminiReq.GenerationConfig.MaxOutputTokens != 1000 {
		t.Errorf("generationConfig max tokens mismatch")
	}
}

func TestConvertGeminiToAnthropicResponse(t *testing.T) {
	geminiResp := &GeminiResponse{
		Candidates: []GeminiCandidate{
			{
				Index:        0,
				FinishReason: "STOP",
				Content: GeminiContent{
					Role: "model",
					Parts: []GeminiPart{
						{
							Text: "The weather in Paris is sunny.",
						},
						{
							FunctionCall: &GeminiFunctionCall{
								Name: "get_weather",
								Args: map[string]interface{}{"location": "Paris"},
							},
							ThoughtSignature: "sig_abc",
						},
					},
				},
			},
		},
		UsageMetadata: &GeminiUsageMetadata{
			PromptTokenCount:     15,
			CandidatesTokenCount: 25,
			ThoughtsTokenCount:   5,
			TotalTokenCount:      45,
		},
	}

	resp := convertGeminiToAnthropicResponse(geminiResp, "gemini-1.5-flash", "test-req-id")
	if resp == nil {
		t.Fatalf("expected non-nil AnthropicResponse")
	}

	if resp.ID != "msg_test-req-id" {
		t.Errorf("expected ID msg_test-req-id, got %s", resp.ID)
	}
	if resp.Model != "gemini-1.5-flash" {
		t.Errorf("expected model gemini-1.5-flash, got %s", resp.Model)
	}
	if resp.Usage.InputTokens != 15 || resp.Usage.OutputTokens != 30 {
		t.Errorf("usage mismatch: %+v", resp.Usage)
	}
	if len(resp.Content) != 2 {
		t.Fatalf("expected 2 content blocks, got %d", len(resp.Content))
	}
	if resp.Content[0].Type != "text" || resp.Content[0].Text != "The weather in Paris is sunny." {
		t.Errorf("text content mismatch: %+v", resp.Content[0])
	}
	if resp.Content[1].Type != "tool_use" || resp.Content[1].Name != "get_weather" {
		t.Errorf("tool_use content mismatch: %+v", resp.Content[1])
	}
	if !strings.Contains(resp.Content[1].ID, "_ts_sig_abc") {
		t.Errorf("expected tool_use ID to contain _ts_sig_abc, got %s", resp.Content[1].ID)
	}
	if resp.StopReason == nil || *resp.StopReason != "tool_use" {
		t.Errorf("expected stop_reason tool_use, got %v", resp.StopReason)
	}
}

func TestConvertAnthropicErrorResponse(t *testing.T) {
	driver, _ := NewAnthropicChat("google-generative-ai", "test-key", "", ai_convert.ModelTypeAnthropicChat, 5*time.Second)
	chat := driver.(*AnthropicChat)

	mockCtx := &mockHttpContext{
		response: &mockResponse{
			statusCode: 400,
		},
	}
	geminiErr := map[string]interface{}{
		"error": map[string]interface{}{
			"code":    400,
			"message": "Invalid argument provided.",
			"status":  "INVALID_ARGUMENT",
		},
	}
	errBytes, _ := json.Marshal(geminiErr)
	mockCtx.response.body = errBytes

	chat.convertErrorResponse(mockCtx, errBytes)

	var anthropicErr struct {
		Type  string `json:"type"`
		Error struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(mockCtx.response.body, &anthropicErr); err != nil {
		t.Fatalf("failed to unmarshal error: %v", err)
	}
	if anthropicErr.Type != "error" || anthropicErr.Error.Message != "Invalid argument provided." {
		t.Errorf("error mismatch: %+v", anthropicErr)
	}
}

func TestAnthropicStreamHandler(t *testing.T) {
	driver, _ := NewAnthropicChat("google-generative-ai", "test-key", "", ai_convert.ModelTypeAnthropicChat, 5*time.Second)
	chat := driver.(*AnthropicChat)

	mockCtx := &mockHttpContext{
		labels:    make(map[string]string),
		requestId: "stream-test-123",
	}
	ai_convert.SetAIModel(mockCtx, "gemini-1.5-flash")

	geminiChunk := `data: {"candidates":[{"index":0,"content":{"role":"model","parts":[{"text":"Hello"}]},"finishReason":""}],"usageMetadata":{"promptTokenCount":10,"candidatesTokenCount":5,"thoughtsTokenCount":0,"totalTokenCount":15}}` + "\n\n"

	out, err := chat.streamHandler(mockCtx, []byte(geminiChunk))
	if err != nil {
		t.Fatalf("streamHandler error: %v", err)
	}
	outStr := string(out)
	if !strings.Contains(outStr, "event: message_start") {
		t.Errorf("expected message_start in output: %s", outStr)
	}
	if !strings.Contains(outStr, "event: content_block_start") {
		t.Errorf("expected content_block_start in output: %s", outStr)
	}
	if !strings.Contains(outStr, "event: content_block_delta") {
		t.Errorf("expected content_block_delta in output: %s", outStr)
	}

	// Terminating chunk
	finalChunk := `data: {"candidates":[{"index":0,"content":{"role":"model","parts":[]},"finishReason":"STOP"}]}` + "\n\n"
	outFinal, err := chat.streamHandler(mockCtx, []byte(finalChunk))
	if err != nil {
		t.Fatalf("streamHandler final chunk error: %v", err)
	}
	outFinalStr := string(outFinal)
	if !strings.Contains(outFinalStr, "event: content_block_stop") {
		t.Errorf("expected content_block_stop in output: %s", outFinalStr)
	}
	if !strings.Contains(outFinalStr, "event: message_delta") {
		t.Errorf("expected message_delta in output: %s", outFinalStr)
	}
	if !strings.Contains(outFinalStr, "event: message_stop") {
		t.Errorf("expected message_stop in output: %s", outFinalStr)
	}
}
