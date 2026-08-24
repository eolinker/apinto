package byteplus

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	ai_convert "github.com/eolinker/apinto/ai-convert"
	context_label "github.com/eolinker/apinto/common/context-label"
	"github.com/eolinker/apinto/encoder"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
	openai "github.com/sashabaranov/go-openai"
)

const (
	labelStreamRemain          = "byteplus_anthropic_stream_remain"
	labelStreamStarted         = "byteplus_anthropic_stream_started"
	labelStreamTextStarted     = "byteplus_anthropic_stream_text_started"
	labelStreamActiveToolIndex = "byteplus_anthropic_stream_active_tool_idx"
	labelStreamOutputTokens    = "byteplus_anthropic_stream_output_tokens"
)

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

func NewAnthropicChat(provider string, apikey string, baseUrlStr string, modelType ai_convert.ModelType, timeout time.Duration) (ai_convert.IConverterDriver, error) {
	c := &AnthropicChat{
		provider:  provider,
		apikey:    apikey,
		modelType: modelType,
	}

	if baseUrlStr == "" {
		baseUrlStr = baseUrl
	}

	balanceHandler, err := ai_convert.NewBalanceHandler(apikey, baseUrlStr, timeout)
	if err != nil {
		return nil, err
	}
	c.balanceHandler = balanceHandler

	u, err := url.Parse(baseUrlStr)
	if err != nil {
		return nil, err
	}
	basePath := strings.TrimSuffix(u.Path, "/")
	if basePath == "" {
		basePath = "/api/v3"
	}
	c.path = fmt.Sprintf("%s%s", basePath, ai_convert.OpenAIChatCompletePath)

	return c, nil
}

func (a *AnthropicChat) Provider() string {
	return a.provider
}

func (a *AnthropicChat) ModelType() ai_convert.ModelType {
	return a.modelType
}

// Anthropic Messages API 输入结构
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

// Anthropic Messages API 响应结构
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
	context_label.SetBillingMode(ctx, context_label.BillingModeImmediate)
	context_label.SetModelCompletionTag(ctx)
	httpContext, err := http_service.Assert(ctx)
	if err != nil {
		return err
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

	openaiReq := convertAnthropicToOpenAIRequest(anthropicReq.Config, anthropicReq.Append)
	newBody, err := json.Marshal(openaiReq)
	if err != nil {
		return fmt.Errorf("marshal openai request error: %v", err)
	}

	if a.apikey != "" {
		httpContext.Proxy().Header().SetHeader("Authorization", "Bearer "+a.apikey)
	}
	httpContext.Proxy().URI().SetPath(a.path)
	httpContext.Proxy().Body().SetRaw("application/json", newBody)

	if anthropicReq.Config.Stream {
		httpContext.Proxy().AppendStreamBodyHandle(a.streamHandler)
	} else {
		context_label.SetDisableStream(ctx, true)
	}

	if a.balanceHandler != nil {
		ctx.SetBalance(a.balanceHandler)
	}
	return nil
}

func convertAnthropicToOpenAIRequest(req *AnthropicMessageRequest, appendParams map[string]interface{}) *openai.ChatCompletionRequest {
	out := &openai.ChatCompletionRequest{
		Model:  req.Model,
		Stream: req.Stream,
	}
	if req.MaxTokens > 0 {
		out.MaxTokens = req.MaxTokens
	}
	if req.Temperature != nil {
		out.Temperature = *req.Temperature
	}
	if req.TopP != nil {
		out.TopP = *req.TopP
	}
	if len(req.StopSequences) > 0 {
		out.Stop = req.StopSequences
	}
	if req.Stream {
		out.StreamOptions = &openai.StreamOptions{
			IncludeUsage: true,
		}
	}

	messages := make([]openai.ChatCompletionMessage, 0)
	if sysContent := parseSystemContent(req.System); sysContent != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: sysContent,
		})
	}

	for _, msg := range req.Messages {
		converted := convertAnthropicMessageToOpenAI(msg)
		messages = append(messages, converted...)
	}
	out.Messages = messages

	for _, tool := range req.Tools {
		out.Tools = append(out.Tools, openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  tool.InputSchema,
			},
		})
	}

	if req.ToolChoice != nil {
		out.ToolChoice = convertAnthropicToolChoiceToOpenAI(req.ToolChoice)
	}

	return out
}

func parseSystemContent(system interface{}) string {
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

func convertAnthropicMessageToOpenAI(msg AnthropicMessage) []openai.ChatCompletionMessage {
	switch c := msg.Content.(type) {
	case string:
		return []openai.ChatCompletionMessage{{
			Role:    msg.Role,
			Content: c,
		}}
	case []interface{}:
		var result []openai.ChatCompletionMessage
		var textParts []string
		var multiParts []openai.ChatMessagePart
		var toolCalls []openai.ToolCall

		for _, item := range c {
			blockBytes, _ := json.Marshal(item)
			var block AnthropicContentBlock
			if err := json.Unmarshal(blockBytes, &block); err != nil {
				continue
			}

			switch block.Type {
			case "text":
				textParts = append(textParts, block.Text)
				multiParts = append(multiParts, openai.ChatMessagePart{
					Type: openai.ChatMessagePartTypeText,
					Text: block.Text,
				})
			case "image":
				if block.Source != nil {
					imgURL := block.Source.URL
					if block.Source.Type == "base64" && block.Source.Data != "" {
						mediaType := block.Source.MediaType
						if mediaType == "" {
							mediaType = "image/jpeg"
						}
						imgURL = fmt.Sprintf("data:%s;base64,%s", mediaType, block.Source.Data)
					}
					if imgURL != "" {
						multiParts = append(multiParts, openai.ChatMessagePart{
							Type: openai.ChatMessagePartTypeImageURL,
							ImageURL: &openai.ChatMessageImageURL{
								URL: imgURL,
							},
						})
					}
				}
			case "tool_use":
				argsBytes, _ := json.Marshal(block.Input)
				toolCalls = append(toolCalls, openai.ToolCall{
					ID:   block.ID,
					Type: openai.ToolTypeFunction,
					Function: openai.FunctionCall{
						Name:      block.Name,
						Arguments: string(argsBytes),
					},
				})
			case "tool_result":
				contentStr := parseContentToString(block.Content)
				result = append(result, openai.ChatCompletionMessage{
					Role:       "tool",
					ToolCallID: block.ToolUseID,
					Content:    contentStr,
				})
			}
		}

		if msg.Role == "user" {
			if len(multiParts) > 0 {
				hasImage := false
				for _, part := range multiParts {
					if part.Type == openai.ChatMessagePartTypeImageURL {
						hasImage = true
						break
					}
				}
				if hasImage {
					result = append(result, openai.ChatCompletionMessage{
						Role:         msg.Role,
						MultiContent: multiParts,
					})
				} else if len(textParts) > 0 {
					result = append(result, openai.ChatCompletionMessage{
						Role:    msg.Role,
						Content: strings.Join(textParts, "\n"),
					})
				}
			}
		} else if msg.Role == "assistant" {
			assistantMsg := openai.ChatCompletionMessage{
				Role: msg.Role,
			}
			if len(textParts) > 0 {
				assistantMsg.Content = strings.Join(textParts, "\n")
			}
			if len(toolCalls) > 0 {
				assistantMsg.ToolCalls = toolCalls
			}
			if assistantMsg.Content != "" || len(assistantMsg.ToolCalls) > 0 {
				result = append(result, assistantMsg)
			}
		}
		return result
	}
	return []openai.ChatCompletionMessage{{
		Role:    msg.Role,
		Content: parseContentToString(msg.Content),
	}}
}

func parseContentToString(content interface{}) string {
	if content == nil {
		return ""
	}
	switch v := content.(type) {
	case string:
		return v
	case []interface{}:
		var builder strings.Builder
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				if t, ok := m["text"].(string); ok {
					builder.WriteString(t)
				}
			}
		}
		return builder.String()
	default:
		b, _ := json.Marshal(v)
		return string(b)
	}
}

func convertAnthropicToolChoiceToOpenAI(toolChoice interface{}) interface{} {
	b, err := json.Marshal(toolChoice)
	if err != nil {
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		switch s {
		case "auto":
			return "auto"
		case "any":
			return "required"
		}
		return nil
	}
	var obj struct {
		Type string `json:"type"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(b, &obj); err == nil {
		switch obj.Type {
		case "auto":
			return "auto"
		case "any":
			return "required"
		case "tool":
			if obj.Name != "" {
				return openai.ToolChoice{
					Type: openai.ToolTypeFunction,
					Function: openai.ToolFunction{
						Name: obj.Name,
					},
				}
			}
		}
	}
	return nil
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
		errorCallback(httpContext, body)
		status := ai_convert.GetAIStatus(ctx)
		if status == "" {
			status = ai_convert.StatusInvalid
		}
		ai_convert.SetAIProviderStatuses(httpContext, status)
		a.convertErrorResponse(httpContext, body)
		return nil
	}

	var openaiResp openai.ChatCompletionResponse
	if err = json.Unmarshal(body, &openaiResp); err != nil {
		ai_convert.SetAIProviderStatuses(httpContext, ai_convert.StatusInvalid)
		log.Errorf("unmarshal openai body error: %v, body: %s", err, string(body))
		return err
	}

	ai_convert.SetAIModelInputToken(httpContext, openaiResp.Usage.PromptTokens)
	ai_convert.SetAIModelOutputToken(httpContext, openaiResp.Usage.CompletionTokens)
	ai_convert.SetAIModelTotalToken(httpContext, openaiResp.Usage.TotalTokens)
	ai_convert.SetAIStatusNormal(ctx)
	ai_convert.SetAIProviderStatuses(httpContext, ai_convert.GetAIStatus(ctx))

	anthropicResp := convertOpenAIToAnthropicResponse(&openaiResp, ai_convert.GetAIModel(ctx))
	newBody, err := json.Marshal(anthropicResp)
	if err != nil {
		return err
	}
	httpContext.Response().SetHeader("content-encoding", "utf-8")
	httpContext.Response().SetBody(newBody)
	return nil
}

func convertOpenAIToAnthropicResponse(resp *openai.ChatCompletionResponse, defaultModel string) *AnthropicResponse {
	model := resp.Model
	if model == "" {
		model = defaultModel
	}
	out := &AnthropicResponse{
		ID:    resp.ID,
		Type:  "message",
		Role:  "assistant",
		Model: model,
		Usage: AnthropicUsage{
			InputTokens:  resp.Usage.PromptTokens,
			OutputTokens: resp.Usage.CompletionTokens,
		},
		Content: make([]AnthropicContentBlock, 0),
	}

	if len(resp.Choices) > 0 {
		choice := resp.Choices[0]
		if choice.Message.Content != "" {
			out.Content = append(out.Content, AnthropicContentBlock{
				Type: "text",
				Text: choice.Message.Content,
			})
		}
		for _, tc := range choice.Message.ToolCalls {
			var input map[string]interface{}
			_ = json.Unmarshal([]byte(tc.Function.Arguments), &input)
			out.Content = append(out.Content, AnthropicContentBlock{
				Type:  "tool_use",
				ID:    tc.ID,
				Name:  tc.Function.Name,
				Input: input,
			})
		}
		reason := mapOpenAIFinishReasonToAnthropic(choice.FinishReason, len(choice.Message.ToolCalls) > 0)
		out.StopReason = &reason
	}

	return out
}

func mapOpenAIFinishReasonToAnthropic(reason openai.FinishReason, hasToolCalls bool) string {
	switch reason {
	case openai.FinishReasonStop:
		return "end_turn"
	case openai.FinishReasonLength:
		return "max_tokens"
	case openai.FinishReasonToolCalls:
		return "tool_use"
	case openai.FinishReasonContentFilter:
		return "end_turn"
	default:
		if hasToolCalls {
			return "tool_use"
		}
		return "end_turn"
	}
}

func (a *AnthropicChat) convertErrorResponse(httpContext http_service.IHttpContext, body []byte) {
	var openaiErr struct {
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Code    string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &openaiErr); err != nil || openaiErr.Error.Message == "" {
		return
	}
	errType := "invalid_request_error"
	if httpContext.Response().StatusCode() == 401 {
		errType = "authentication_error"
	} else if httpContext.Response().StatusCode() == 429 {
		errType = "rate_limit_error"
	} else if httpContext.Response().StatusCode() >= 500 {
		errType = "api_error"
	}
	anthropicErr := map[string]interface{}{
		"type": "error",
		"error": map[string]interface{}{
			"type":    errType,
			"message": openaiErr.Error.Message,
		},
	}
	newBody, err := json.Marshal(anthropicErr)
	if err != nil {
		return
	}
	httpContext.Response().SetBody(newBody)
}

func (a *AnthropicChat) streamHandler(ctx http_service.IHttpContext, p []byte) ([]byte, error) {
	var sseBuffer bytes.Buffer
	model := ai_convert.GetAIModel(ctx)
	requestID := ctx.RequestId()

	data := ctx.GetLabel(labelStreamRemain) + string(p)
	lastNL := strings.LastIndexByte(data, '\n')
	if lastNL < 0 {
		ctx.SetLabel(labelStreamRemain, data)
		return []byte{}, nil
	}
	ctx.SetLabel(labelStreamRemain, data[lastNL+1:])
	complete := data[:lastNL+1]

	scanner := bufio.NewScanner(strings.NewReader(complete))
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)

	started := ctx.GetLabel(labelStreamStarted) == "true"
	textStarted := ctx.GetLabel(labelStreamTextStarted) == "true"
	outTokens := 0
	fmt.Sscanf(ctx.GetLabel(labelStreamOutputTokens), "%d", &outTokens)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || !strings.HasPrefix(line, "data:") {
			continue
		}
		dataStr := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if dataStr == "" {
			continue
		}

		if dataStr == "[DONE]" {
			if textStarted {
				writeAnthropicSSE(&sseBuffer, "content_block_stop", map[string]interface{}{
					"type":  "content_block_stop",
					"index": 0,
				})
				textStarted = false
				ctx.SetLabel(labelStreamTextStarted, "false")
			}
			stopReason := "end_turn"
			writeAnthropicSSE(&sseBuffer, "message_delta", map[string]interface{}{
				"type": "message_delta",
				"delta": map[string]interface{}{
					"stop_reason":   stopReason,
					"stop_sequence": nil,
				},
				"usage": map[string]interface{}{
					"output_tokens": outTokens,
				},
			})
			writeAnthropicSSE(&sseBuffer, "message_stop", map[string]interface{}{
				"type": "message_stop",
			})
			continue
		}

		var chunk openai.ChatCompletionStreamResponse
		if err := json.Unmarshal([]byte(dataStr), &chunk); err != nil {
			continue
		}

		if chunk.ID != "" {
			requestID = chunk.ID
		}
		if chunk.Model != "" {
			model = chunk.Model
		}

		if !started {
			started = true
			ctx.SetLabel(labelStreamStarted, "true")
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

		if chunk.Usage != nil {
			ai_convert.SetAIModelInputToken(ctx, chunk.Usage.PromptTokens)
			ai_convert.SetAIModelOutputToken(ctx, chunk.Usage.CompletionTokens)
			ai_convert.SetAIModelTotalToken(ctx, chunk.Usage.TotalTokens)
			outTokens = chunk.Usage.CompletionTokens
			ctx.SetLabel(labelStreamOutputTokens, fmt.Sprintf("%d", outTokens))
		}

		if len(chunk.Choices) > 0 {
			choice := chunk.Choices[0]
			if choice.Delta.Content != "" {
				if !textStarted {
					textStarted = true
					ctx.SetLabel(labelStreamTextStarted, "true")
					writeAnthropicSSE(&sseBuffer, "content_block_start", map[string]interface{}{
						"type":  "content_block_start",
						"index": 0,
						"content_block": map[string]interface{}{
							"type": "text",
							"text": "",
						},
					})
				}
				outTokens++
				ctx.SetLabel(labelStreamOutputTokens, fmt.Sprintf("%d", outTokens))
				writeAnthropicSSE(&sseBuffer, "content_block_delta", map[string]interface{}{
					"type":  "content_block_delta",
					"index": 0,
					"delta": map[string]interface{}{
						"type": "text_delta",
						"text": choice.Delta.Content,
					},
				})
			}

			for _, tc := range choice.Delta.ToolCalls {
				tcIndex := 1
				if tc.Index != nil {
					tcIndex = *tc.Index + 1
				}
				if tc.ID != "" || tc.Function.Name != "" {
					writeAnthropicSSE(&sseBuffer, "content_block_start", map[string]interface{}{
						"type":  "content_block_start",
						"index": tcIndex,
						"content_block": map[string]interface{}{
							"type":  "tool_use",
							"id":    tc.ID,
							"name":  tc.Function.Name,
							"input": map[string]interface{}{},
						},
					})
				}
				if tc.Function.Arguments != "" {
					writeAnthropicSSE(&sseBuffer, "content_block_delta", map[string]interface{}{
						"type":  "content_block_delta",
						"index": tcIndex,
						"delta": map[string]interface{}{
							"type":         "input_json_delta",
							"partial_json": tc.Function.Arguments,
						},
					})
				}
			}

			if choice.FinishReason != "" {
				if textStarted {
					writeAnthropicSSE(&sseBuffer, "content_block_stop", map[string]interface{}{
						"type":  "content_block_stop",
						"index": 0,
					})
					textStarted = false
					ctx.SetLabel(labelStreamTextStarted, "false")
				}
				stopReason := mapOpenAIFinishReasonToAnthropic(choice.FinishReason, len(choice.Delta.ToolCalls) > 0)
				writeAnthropicSSE(&sseBuffer, "message_delta", map[string]interface{}{
					"type": "message_delta",
					"delta": map[string]interface{}{
						"stop_reason":   stopReason,
						"stop_sequence": nil,
					},
					"usage": map[string]interface{}{
						"output_tokens": outTokens,
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

func writeAnthropicSSE(buf *bytes.Buffer, event string, data interface{}) {
	payload, _ := json.Marshal(data)
	buf.WriteString(fmt.Sprintf("event: %s\ndata: %s\n\n", event, string(payload)))
}
