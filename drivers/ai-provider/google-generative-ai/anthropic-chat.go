package google

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	ai_convert "github.com/eolinker/apinto/ai-convert"
	context_label2 "github.com/eolinker/apinto/common/context-label"
	"github.com/eolinker/apinto/encoder"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
)

type anthropicStreamState struct {
	buffer               []byte
	lastThoughtSignature string
	started              bool
	textStarted          bool
	outTokens            int
	toolIndex            int
}

const anthropicStreamStateKey = "google_anthropic_stream_state"

func getAnthropicStreamState(ctx http_service.IHttpContext) *anthropicStreamState {
	if val := ctx.Value(anthropicStreamStateKey); val != nil {
		if state, ok := val.(*anthropicStreamState); ok {
			return state
		}
	}
	state := &anthropicStreamState{}
	ctx.WithValue(anthropicStreamStateKey, state)
	return state
}

func init() {
	driverCreate.Set(ai_convert.ModelTypeAnthropicChat, func(mt ai_convert.ModelType, c *Config) (ai_convert.IConverterDriver, error) {
		return NewAnthropicChat(provider, c.APIKey, c.BaseUrl, mt, 10*time.Minute)
	})
}

var _ ai_convert.IConverterDriver = (*AnthropicChat)(nil)

type AnthropicChat struct {
	provider       string
	apikey         string
	path           string
	modelType      ai_convert.ModelType
	balanceHandler eocontext.BalanceHandler
}

func NewAnthropicChat(provider string, apikey string, baseUrl string, modelType ai_convert.ModelType, timeout time.Duration) (ai_convert.IConverterDriver, error) {
	c := &AnthropicChat{
		provider:  provider,
		apikey:    apikey,
		modelType: modelType,
	}
	orgBase := generativeAIBase
	orgPath := generativeAIPath
	if baseUrl == "" {
		baseUrl = orgBase
	}

	balanceHandler, err := ai_convert.NewBalanceHandler(apikey, baseUrl, timeout)
	if err != nil {
		return nil, err
	}
	c.balanceHandler = balanceHandler
	u, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}
	if strings.TrimSuffix(u.Path, "/") == "" {
		c.path = "/v1beta" + orgPath
	} else {
		c.path = fmt.Sprintf("%s%s", strings.TrimSuffix(u.Path, "/"), orgPath)
	}
	return c, nil
}

func (a *AnthropicChat) Provider() string {
	return a.provider
}

func (a *AnthropicChat) ModelType() ai_convert.ModelType {
	return a.modelType
}

// Anthropic Messages API Request structures
type AnthropicMessageRequest struct {
	Model         string             `json:"model"`
	Messages      []AnthropicMessage `json:"messages"`
	System        interface{}        `json:"system,omitempty"`
	MaxTokens     int                `json:"max_tokens"`
	Stream        bool               `json:"stream,omitempty"`
	Temperature   *float32           `json:"temperature,omitempty"`
	TopP          *float32           `json:"top_p,omitempty"`
	StopSequences []string           `json:"stop_sequences,omitempty"`
	Tools         []AnthropicTool    `json:"tools,omitempty"`
	ToolChoice    interface{}        `json:"tool_choice,omitempty"`
}

type AnthropicMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

type AnthropicContentBlock struct {
	Type      string                 `json:"type"`
	Text      string                 `json:"text,omitempty"`
	Source    *AnthropicImageSource  `json:"source,omitempty"`
	ID        string                 `json:"id,omitempty"`
	Name      string                 `json:"name,omitempty"`
	Input     map[string]interface{} `json:"input,omitempty"`
	ToolUseID string                 `json:"tool_use_id,omitempty"`
	Content   interface{}            `json:"content,omitempty"`
}

type AnthropicImageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type,omitempty"`
	Data      string `json:"data,omitempty"`
	URL       string `json:"url,omitempty"`
}

type AnthropicTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"input_schema,omitempty"`
}

// Anthropic Messages API Response structures
type AnthropicResponse struct {
	ID         string                  `json:"id"`
	Type       string                  `json:"type"`
	Role       string                  `json:"role"`
	Model      string                  `json:"model"`
	Content    []AnthropicContentBlock `json:"content"`
	StopReason *string                 `json:"stop_reason"`
	Usage      AnthropicUsage          `json:"usage"`
}

type AnthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

func (a *AnthropicChat) RequestConvert(ctx eocontext.EoContext, extender map[string]interface{}) error {
	context_label2.SetBillingMode(ctx, context_label2.BillingModeImmediate)
	httpContext, err := http_service.Assert(ctx)
	if err != nil {
		return err
	}
	if a.apikey != "" {
		httpContext.Proxy().Header().SetHeader("x-goog-api-key", a.apikey)
		httpContext.Proxy().Header().DelHeader("authorization")
	}

	body, err := httpContext.Proxy().Body().RawBody()
	if err != nil {
		return err
	}

	anthropicReq := eosc.NewBase[AnthropicMessageRequest](extender)
	if err = json.Unmarshal(body, anthropicReq); err != nil {
		return fmt.Errorf("unmarshal anthropic request error: %v, body: %s", err, string(body))
	}
	if anthropicReq.Config.Model == "" {
		anthropicReq.Config.Model = ai_convert.GetAIModel(ctx)
	}
	model := anthropicReq.Config.Model

	geminiReq := convertAnthropicToGeminiRequest(anthropicReq.Config)

	newBody, err := json.Marshal(geminiReq)
	if err != nil {
		return fmt.Errorf("marshal gemini request error: %v", err)
	}
	httpContext.Proxy().Body().SetRaw("application/json", newBody)
	httpContext.Proxy().URI().DelQuery("beta")
	// Set URL and Stream handling
	if anthropicReq.Config.Stream {
		path := fmt.Sprintf("%s/%s:%s", a.path, model, "streamGenerateContent")
		httpContext.Proxy().URI().SetPath(path)
		httpContext.Proxy().URI().SetQuery("alt", "sse")
		httpContext.Proxy().AppendStreamBodyHandle(a.streamHandler)
		httpContext.Proxy().AppendBodyFinish(a.streamFinish)
		context_label2.SetModelCompletionStreamTag(ctx)
	} else {
		path := fmt.Sprintf("%s/%s:%s", a.path, model, "generateContent")
		httpContext.Proxy().URI().SetPath(path)
		context_label2.SetDisableStream(ctx, true)
		context_label2.SetModelCompletionTag(ctx)
	}

	if a.balanceHandler != nil {
		ctx.SetBalance(a.balanceHandler)
	}

	return nil
}

func convertAnthropicToGeminiRequest(req *AnthropicMessageRequest) *GeminiRequest {
	geminiReq := &GeminiRequest{}

	// 1. System Instruction
	if sysText := parseAnthropicSystem(req.System); sysText != "" {
		geminiReq.SystemInstruction = &GeminiContent{
			Role: "user",
			Parts: []GeminiPart{
				{Text: sysText},
			},
		}
	}

	// 2. Messages -> Contents
	for _, msg := range req.Messages {
		content := GeminiContent{}
		if msg.Role == "assistant" {
			content.Role = "model"
		} else {
			content.Role = "user"
		}
		var lastValidSig string

		switch c := msg.Content.(type) {
		case string:
			if c != "" {
				content.Parts = append(content.Parts, GeminiPart{Text: c})
			}
		case []interface{}:
			for _, item := range c {
				blockBytes, err := json.Marshal(item)
				if err != nil {
					continue
				}
				var block AnthropicContentBlock
				if err := json.Unmarshal(blockBytes, &block); err != nil {
					continue
				}

				switch block.Type {
				case "text":
					if block.Text != "" {
						content.Parts = append(content.Parts, GeminiPart{Text: block.Text})
					}
				case "image":
					if block.Source != nil {
						if block.Source.Type == "base64" && block.Source.Data != "" {
							mediaType := block.Source.MediaType
							if mediaType == "" {
								mediaType = "image/jpeg"
							}
							content.Parts = append(content.Parts, GeminiPart{
								InlineData: &GeminiInlineData{
									MimeType: mediaType,
									Data:     block.Source.Data,
								},
							})
						} else if block.Source.URL != "" {
							content.Parts = append(content.Parts, GeminiPart{
								FileData: &GeminiFileData{
									MimeType: mimeTypeFromURL(block.Source.URL),
									FileURI:  block.Source.URL,
								},
							})
						}
					}
				case "tool_use":
					ts := ""
					if strings.Contains(block.ID, "_ts_") {
						parts := strings.SplitN(block.ID, "_ts_", 2)
						if len(parts) == 2 {
							ts = parts[1]
						}
					}
					if ts == "" {
						ts = lookupCachedThoughtSignature(block.Name, block.Input)
					}
					if ts != "" {
						lastValidSig = ts
					}
					content.Parts = append(content.Parts, GeminiPart{
						FunctionCall: &GeminiFunctionCall{
							Name: block.Name,
							Args: block.Input,
						},
						ThoughtSignature: ts,
					})
				case "tool_result":
					var responseObj map[string]interface{}
					switch rc := block.Content.(type) {
					case string:
						if err := json.Unmarshal([]byte(rc), &responseObj); err != nil {
							responseObj = map[string]interface{}{"result": rc}
						}
					case map[string]interface{}:
						responseObj = rc
					case []interface{}:
						contentStr := parseAnthropicContentBlocksToString(rc)
						if err := json.Unmarshal([]byte(contentStr), &responseObj); err != nil {
							responseObj = map[string]interface{}{"result": contentStr}
						}
					default:
						if b, err := json.Marshal(rc); err == nil {
							if err := json.Unmarshal(b, &responseObj); err != nil {
								responseObj = map[string]interface{}{"result": string(b)}
							}
						}
					}

					sanitized := make(map[string]interface{}, len(responseObj))
					for k, v := range responseObj {
						if strings.HasPrefix(k, "$") {
							continue
						}
						sanitized[k] = sanitizeFunctionResponse(v)
					}
					if len(sanitized) == 0 && len(responseObj) > 0 {
						sanitized["result"] = responseObj
					}
					responseObj = sanitized

					funcName := block.Name
					if funcName == "" && block.ToolUseID != "" {
						// Look backward to find tool name by ToolUseID in assistant messages
						for i := len(req.Messages) - 1; i >= 0; i-- {
							prevMsg := req.Messages[i]
							if prevMsg.Role == "assistant" {
								if blocks, ok := prevMsg.Content.([]interface{}); ok {
									for _, b := range blocks {
										bb, _ := json.Marshal(b)
										var prevBlock AnthropicContentBlock
										if err := json.Unmarshal(bb, &prevBlock); err == nil {
											if prevBlock.Type == "tool_use" && prevBlock.ID == block.ToolUseID {
												funcName = prevBlock.Name
												break
											}
										}
									}
								}
							}
							if funcName != "" {
								break
							}
						}
					}
					if funcName == "" {
						if block.ToolUseID != "" {
							funcName = block.ToolUseID
						} else {
							funcName = "unknown_function"
						}
					}

					content.Parts = append(content.Parts, GeminiPart{
						FunctionResponse: &GeminiFunctionResponse{
							Name:     funcName,
							Response: responseObj,
						},
					})
				}
			}
		}

		// If some function calls missed signatures, backfill with lastValidSig
		if lastValidSig != "" {
			for i := range content.Parts {
				if content.Parts[i].FunctionCall != nil && content.Parts[i].ThoughtSignature == "" {
					content.Parts[i].ThoughtSignature = lastValidSig
				}
			}
		}

		if len(content.Parts) > 0 {
			geminiReq.Contents = append(geminiReq.Contents, content)
		}
	}

	// 3. Tools
	if len(req.Tools) > 0 {
		var declarations []GeminiFunctionDeclaration
		for _, tool := range req.Tools {
			decl := GeminiFunctionDeclaration{
				Name:        tool.Name,
				Description: tool.Description,
			}
			if tool.InputSchema != nil {
				if paramBytes, err := json.Marshal(tool.InputSchema); err == nil {
					var paramMap map[string]interface{}
					if err := json.Unmarshal(paramBytes, &paramMap); err == nil {
						convertSchemaToGemini(paramMap)
						decl.Parameters = paramMap
					}
				}
			}
			declarations = append(declarations, decl)
		}
		if len(declarations) > 0 {
			geminiReq.Tools = []GeminiTool{{FunctionDeclarations: declarations}}
		}
	}

	// 4. Tool Choice
	if req.ToolChoice != nil {
		config := &GeminiToolConfig{
			FunctionCallingConfig: &GeminiFunctionCallingConfig{},
		}
		var toolChoiceStr string
		if bytes, err := json.Marshal(req.ToolChoice); err == nil {
			if err := json.Unmarshal(bytes, &toolChoiceStr); err == nil {
				switch toolChoiceStr {
				case "auto":
					config.FunctionCallingConfig.Mode = "AUTO"
				case "any":
					config.FunctionCallingConfig.Mode = "ANY"
				case "none":
					config.FunctionCallingConfig.Mode = "NONE"
				}
			} else {
				var toolChoiceObj struct {
					Type string `json:"type"`
					Name string `json:"name"`
				}
				if err := json.Unmarshal(bytes, &toolChoiceObj); err == nil {
					switch toolChoiceObj.Type {
					case "auto":
						config.FunctionCallingConfig.Mode = "AUTO"
					case "any":
						config.FunctionCallingConfig.Mode = "ANY"
					case "tool":
						if toolChoiceObj.Name != "" {
							config.FunctionCallingConfig.Mode = "ANY"
							config.FunctionCallingConfig.AllowedFunctionNames = []string{toolChoiceObj.Name}
						}
					}
				}
			}
		}
		if config.FunctionCallingConfig.Mode != "" {
			geminiReq.ToolConfig = config
		}
	}

	// 5. Generation Config
	genConfig := &GeminiGenerationConfig{}
	if req.Temperature != nil {
		genConfig.Temperature = req.Temperature
	}
	if req.TopP != nil {
		genConfig.TopP = req.TopP
	}
	if req.MaxTokens > 0 {
		genConfig.MaxOutputTokens = &req.MaxTokens
	}
	if len(req.StopSequences) > 0 {
		genConfig.StopSequences = req.StopSequences
	}
	geminiReq.GenerationConfig = genConfig

	return geminiReq
}

func parseAnthropicSystem(system interface{}) string {
	if system == nil {
		return ""
	}
	switch s := system.(type) {
	case string:
		return s
	case []interface{}:
		var builder strings.Builder
		for _, item := range s {
			if m, ok := item.(map[string]interface{}); ok {
				if t, ok := m["text"].(string); ok {
					if builder.Len() > 0 {
						builder.WriteString("\n")
					}
					builder.WriteString(t)
				}
			}
		}
		return builder.String()
	}
	return ""
}

func parseAnthropicContentBlocksToString(blocks []interface{}) string {
	var builder strings.Builder
	for _, item := range blocks {
		if m, ok := item.(map[string]interface{}); ok {
			if t, ok := m["text"].(string); ok {
				builder.WriteString(t)
			}
		}
	}
	return builder.String()
}

func (a *AnthropicChat) ResponseConvert(ctx eocontext.EoContext) error {
	httpContext, err := http_service.Assert(ctx)
	if err != nil {
		return err
	}
	body := httpContext.Response().GetBody()
	encoding := httpContext.Response().Headers().Get("content-encoding")
	if encoding != "utf-8" && encoding != "" {
		body, err = encoder.ToUTF8(encoding, body)
		if err != nil {
			return err
		}
	}

	if httpContext.Response().StatusCode() != 200 {
		status := ai_convert.GetAIStatus(ctx)
		if status == "" {
			status = ai_convert.StatusInvalid
		}
		ai_convert.SetAIProviderStatuses(httpContext, status)
		a.convertErrorResponse(httpContext, body)
		return nil
	}

	var geminiResp GeminiResponse
	if err = json.Unmarshal(body, &geminiResp); err != nil {
		ai_convert.SetAIProviderStatuses(httpContext, ai_convert.StatusInvalid)
		log.Errorf("unmarshal gemini response error: %v, body: %s", err, string(body))
		return err
	}

	if geminiResp.UsageMetadata != nil {
		outputTokens := geminiResp.UsageMetadata.CandidatesTokenCount + geminiResp.UsageMetadata.ThoughtsTokenCount
		ai_convert.SetAIModelInputToken(httpContext, geminiResp.UsageMetadata.PromptTokenCount)
		ai_convert.SetAIModelOutputToken(httpContext, outputTokens)
		ai_convert.SetAIModelTotalToken(httpContext, geminiResp.UsageMetadata.TotalTokenCount)
	}
	ai_convert.SetAIStatusNormal(ctx)
	ai_convert.SetAIProviderStatuses(httpContext, ai_convert.GetAIStatus(ctx))

	anthropicResp := convertGeminiToAnthropicResponse(&geminiResp, ai_convert.GetAIModel(ctx), ctx.RequestId())
	newBody, err := json.Marshal(anthropicResp)
	if err != nil {
		return err
	}
	httpContext.Response().SetHeader("content-encoding", "utf-8")
	httpContext.Response().SetBody(newBody)
	return nil
}

func convertGeminiToAnthropicResponse(resp *GeminiResponse, defaultModel string, reqID string) *AnthropicResponse {
	out := &AnthropicResponse{
		ID:      "msg_" + reqID,
		Type:    "message",
		Role:    "assistant",
		Model:   defaultModel,
		Content: make([]AnthropicContentBlock, 0),
	}

	if resp.UsageMetadata != nil {
		out.Usage = AnthropicUsage{
			InputTokens:  resp.UsageMetadata.PromptTokenCount,
			OutputTokens: resp.UsageMetadata.CandidatesTokenCount + resp.UsageMetadata.ThoughtsTokenCount,
		}
	}

	if len(resp.Candidates) > 0 {
		candidate := resp.Candidates[0]
		hasToolCall := false
		lastSig := ""
		for _, part := range candidate.Content.Parts {
			if part.ThoughtSignature != "" {
				lastSig = part.ThoughtSignature
			}
		}
		for _, part := range candidate.Content.Parts {
			if part.Text != "" {
				out.Content = append(out.Content, AnthropicContentBlock{
					Type: "text",
					Text: part.Text,
				})
			}
			if part.FunctionCall != nil {
				hasToolCall = true
				sig := part.ThoughtSignature
				if sig == "" {
					sig = lastSig
				}
				toolCallID := formatToolCallID(sig)
				if sig != "" {
					cacheThoughtSignature(part.FunctionCall.Name, part.FunctionCall.Args, sig)
				}
				out.Content = append(out.Content, AnthropicContentBlock{
					Type:  "tool_use",
					ID:    toolCallID,
					Name:  part.FunctionCall.Name,
					Input: part.FunctionCall.Args,
				})
			}
		}

		stopReason := mapGeminiFinishReasonToAnthropic(candidate.FinishReason, hasToolCall)
		out.StopReason = &stopReason
	}

	return out
}

func mapGeminiFinishReasonToAnthropic(reason string, hasToolCalls bool) string {
	switch reason {
	case "STOP":
		if hasToolCalls {
			return "tool_use"
		}
		return "end_turn"
	case "MAX_TOKENS":
		return "max_tokens"
	case "SAFETY", "RECITATION":
		return "end_turn"
	default:
		if hasToolCalls {
			return "tool_use"
		}
		return "end_turn"
	}
}

func (a *AnthropicChat) convertErrorResponse(httpContext http_service.IHttpContext, body []byte) {
	var geminiErr struct {
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Status  string `json:"status"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &geminiErr); err != nil || geminiErr.Error.Message == "" {
		return
	}

	errType := "invalid_request_error"
	if httpContext.Response().StatusCode() == 401 || geminiErr.Error.Code == 401 {
		errType = "authentication_error"
	} else if httpContext.Response().StatusCode() == 429 || geminiErr.Error.Code == 429 {
		errType = "rate_limit_error"
	} else if httpContext.Response().StatusCode() >= 500 || geminiErr.Error.Code >= 500 {
		errType = "api_error"
	}

	anthropicErr := map[string]interface{}{
		"type": "error",
		"error": map[string]interface{}{
			"type":    errType,
			"message": geminiErr.Error.Message,
		},
	}
	newBody, err := json.Marshal(anthropicErr)
	if err != nil {
		return
	}
	httpContext.Response().SetBody(newBody)
}

func (a *AnthropicChat) streamHandler(ctx http_service.IHttpContext, p []byte) ([]byte, error) {
	state := getAnthropicStreamState(ctx)
	state.buffer = append(state.buffer, p...)

	lastNL := bytes.LastIndexByte(state.buffer, '\n')
	if lastNL == -1 {
		return []byte{}, nil
	}

	completeData := state.buffer[:lastNL]
	state.buffer = append([]byte(nil), state.buffer[lastNL+1:]...)

	var sseBuffer bytes.Buffer
	requestID := ctx.RequestId()
	model := ai_convert.GetAIModel(ctx)

	lines := bytes.Split(completeData, []byte("\n"))
	for _, lineBytes := range lines {
		lineBytes = bytes.TrimSpace(lineBytes)
		if len(lineBytes) == 0 || !bytes.HasPrefix(lineBytes, []byte("data:")) {
			continue
		}
		dataBytes := bytes.TrimSpace(bytes.TrimPrefix(lineBytes, []byte("data:")))
		if len(dataBytes) == 0 {
			continue
		}

		if bytes.Equal(dataBytes, []byte("[DONE]")) {
			if state.textStarted {
				writeAnthropicSSE(&sseBuffer, "content_block_stop", map[string]interface{}{
					"type":  "content_block_stop",
					"index": 0,
				})
				state.textStarted = false
			}
			writeAnthropicSSE(&sseBuffer, "message_delta", map[string]interface{}{
				"type": "message_delta",
				"delta": map[string]interface{}{
					"stop_reason":   "end_turn",
					"stop_sequence": nil,
				},
				"usage": map[string]interface{}{
					"output_tokens": state.outTokens,
				},
			})
			writeAnthropicSSE(&sseBuffer, "message_stop", map[string]interface{}{
				"type": "message_stop",
			})
			continue
		}

		var geminiResp GeminiResponse
		if err := json.Unmarshal(dataBytes, &geminiResp); err != nil {
			log.Errorf("unmarshal gemini stream chunk error: %v, data: %s", err, string(dataBytes))
			continue
		}

		if geminiResp.UsageMetadata != nil {
			ai_convert.SetAIModelInputToken(ctx, geminiResp.UsageMetadata.PromptTokenCount)
			outputTokens := geminiResp.UsageMetadata.CandidatesTokenCount + geminiResp.UsageMetadata.ThoughtsTokenCount
			ai_convert.SetAIModelOutputToken(ctx, outputTokens)
			ai_convert.SetAIModelTotalToken(ctx, geminiResp.UsageMetadata.TotalTokenCount)
			state.outTokens = outputTokens
		}

		for _, candidate := range geminiResp.Candidates {
			for _, part := range candidate.Content.Parts {
				if part.ThoughtSignature != "" {
					state.lastThoughtSignature = part.ThoughtSignature
				}
			}

			if !state.started {
				state.started = true
				writeAnthropicSSE(&sseBuffer, "message_start", map[string]interface{}{
					"type": "message_start",
					"message": map[string]interface{}{
						"id":            "msg_" + requestID,
						"type":          "message",
						"role":          "assistant",
						"content":       []interface{}{},
						"model":         model,
						"stop_reason":   nil,
						"stop_sequence": nil,
						"usage": map[string]interface{}{
							"input_tokens":  ai_convert.GetAIModelInputToken(ctx),
							"output_tokens": 1,
						},
					},
				})
			}

			hasToolCall := false
			for _, part := range candidate.Content.Parts {
				if part.Text != "" {
					if !state.textStarted {
						state.textStarted = true
						writeAnthropicSSE(&sseBuffer, "content_block_start", map[string]interface{}{
							"type":  "content_block_start",
							"index": 0,
							"content_block": map[string]interface{}{
								"type": "text",
								"text": "",
							},
						})
					}
					state.outTokens++
					writeAnthropicSSE(&sseBuffer, "content_block_delta", map[string]interface{}{
						"type":  "content_block_delta",
						"index": 0,
						"delta": map[string]interface{}{
							"type": "text_delta",
							"text": part.Text,
						},
					})
				}

				if part.FunctionCall != nil {
					hasToolCall = true
					state.toolIndex++

					sig := part.ThoughtSignature
					if sig == "" {
						sig = state.lastThoughtSignature
					}
					toolCallID := formatToolCallID(sig)
					if sig != "" {
						cacheThoughtSignatureWithContext(ctx, part.FunctionCall.Name, part.FunctionCall.Args, sig)
					}

					writeAnthropicSSE(&sseBuffer, "content_block_start", map[string]interface{}{
						"type":  "content_block_start",
						"index": state.toolIndex,
						"content_block": map[string]interface{}{
							"type":  "tool_use",
							"id":    toolCallID,
							"name":  part.FunctionCall.Name,
							"input": map[string]interface{}{},
						},
					})

					argsBytes, _ := json.Marshal(part.FunctionCall.Args)
					writeAnthropicSSE(&sseBuffer, "content_block_delta", map[string]interface{}{
						"type":  "content_block_delta",
						"index": state.toolIndex,
						"delta": map[string]interface{}{
							"type":         "input_json_delta",
							"partial_json": string(argsBytes),
						},
					})

					writeAnthropicSSE(&sseBuffer, "content_block_stop", map[string]interface{}{
						"type":  "content_block_stop",
						"index": state.toolIndex,
					})
				}
			}

			if candidate.FinishReason != "" {
				if state.textStarted {
					writeAnthropicSSE(&sseBuffer, "content_block_stop", map[string]interface{}{
						"type":  "content_block_stop",
						"index": 0,
					})
					state.textStarted = false
				}
				stopReason := mapGeminiFinishReasonToAnthropic(candidate.FinishReason, hasToolCall)
				writeAnthropicSSE(&sseBuffer, "message_delta", map[string]interface{}{
					"type": "message_delta",
					"delta": map[string]interface{}{
						"stop_reason":   stopReason,
						"stop_sequence": nil,
					},
					"usage": map[string]interface{}{
						"output_tokens": state.outTokens,
					},
				})
				writeAnthropicSSE(&sseBuffer, "message_stop", map[string]interface{}{
					"type": "message_stop",
				})
			}
		}
	}

	return sseBuffer.Bytes(), nil
}

func (a *AnthropicChat) streamFinish(ctx http_service.IHttpContext) {
	state := getAnthropicStreamState(ctx)
	if len(state.buffer) == 0 {
		return
	}
	remaining := bytes.TrimSpace(state.buffer)
	state.buffer = nil
	if len(remaining) == 0 {
		return
	}

	lines := bytes.Split(remaining, []byte("\n"))
	for _, lineBytes := range lines {
		lineBytes = bytes.TrimSpace(lineBytes)
		if len(lineBytes) == 0 || !bytes.HasPrefix(lineBytes, []byte("data:")) {
			continue
		}
		dataBytes := bytes.TrimSpace(bytes.TrimPrefix(lineBytes, []byte("data:")))
		if len(dataBytes) == 0 || bytes.Equal(dataBytes, []byte("[DONE]")) {
			continue
		}
		var geminiResp GeminiResponse
		if err := json.Unmarshal(dataBytes, &geminiResp); err != nil {
			continue
		}
		for _, candidate := range geminiResp.Candidates {
			for _, part := range candidate.Content.Parts {
				if part.ThoughtSignature != "" {
					state.lastThoughtSignature = part.ThoughtSignature
				}
			}
			for _, part := range candidate.Content.Parts {
				if part.FunctionCall != nil {
					sig := part.ThoughtSignature
					if sig == "" {
						sig = state.lastThoughtSignature
					}
					if sig != "" {
						cacheThoughtSignatureWithContext(ctx, part.FunctionCall.Name, part.FunctionCall.Args, sig)
					}
				}
			}
		}
	}
}

func writeAnthropicSSE(buf *bytes.Buffer, event string, data interface{}) {
	payload, _ := json.Marshal(data)
	buf.WriteString(fmt.Sprintf("event: %s\ndata: %s\n\n", event, string(payload)))
}
