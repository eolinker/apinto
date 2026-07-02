package anthropic

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	context_label2 "github.com/eolinker/apinto/common/context-label"
	"net/url"
	"strings"
	"time"

	ai_convert "github.com/eolinker/apinto/ai-convert"
	"github.com/eolinker/apinto/encoder"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
	"github.com/sashabaranov/go-openai"
)

func init() {
	driverCreate.Set(ai_convert.ModelTypeOpenAIChat, func(mt ai_convert.ModelType, c *Config) (ai_convert.IConverterDriver, error) {
		return NewOpenAIChat(provider, c.APIKey, c.Base, mt, 10*time.Minute)
	})
}

func errorCallback(ctx http_service.IHttpContext, body []byte) {
	switch ctx.Response().StatusCode() {
	case 400:
		// Handle the bad request error.
		ai_convert.SetAIStatusInvalidRequest(ctx)
	case 429:
		// Handle exceed
		ai_convert.SetAIStatusExceeded(ctx)
	case 401:
		// Handle authentication failure
		ai_convert.SetAIStatusInvalid(ctx)
	}
}

// defaultMaxTokens Anthropic requires max_tokens; use a sane default when the
// client does not provide one.
const defaultMaxTokens = 4096

func NewOpenAIChat(provider string, apikey string, baseUrl string, modelType ai_convert.ModelType, timeout time.Duration) (ai_convert.IConverterDriver, error) {
	c := &OpenAIChat{
		provider:  provider,
		apikey:    apikey,
		modelType: modelType,
	}
	if baseUrl != "" {
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
			u.Path = "/v1"
		}
		c.path = fmt.Sprintf("%s%s", strings.TrimSuffix(u.Path, "/"), chatPath)
	} else {
		c.path = fmt.Sprintf("/v1%s", chatPath)
	}
	return c, nil
}

type OpenAIChat struct {
	provider       string
	apikey         string
	path           string
	modelType      ai_convert.ModelType
	balanceHandler eocontext.BalanceHandler
}

func (o *OpenAIChat) Provider() string {
	return o.provider
}

func (o *OpenAIChat) ModelType() ai_convert.ModelType {
	return o.modelType
}

// ---- Anthropic Messages API structures ----

type anthropicRequest struct {
	Model         string             `json:"model"`
	MaxTokens     int                `json:"max_tokens"`
	Messages      []anthropicMessage `json:"messages"`
	System        string             `json:"system,omitempty"`
	Stream        bool               `json:"stream,omitempty"`
	Temperature   *float32           `json:"temperature,omitempty"`
	TopP          *float32           `json:"top_p,omitempty"`
	StopSequences []string           `json:"stop_sequences,omitempty"`
	Tools         []anthropicTool    `json:"tools,omitempty"`
	ToolChoice    interface{}        `json:"tool_choice,omitempty"`
}

type anthropicMessage struct {
	Role    string                  `json:"role"`
	Content []anthropicContentBlock `json:"content"`
}

type anthropicContentBlock struct {
	Type string `json:"type"`
	// text block
	Text string `json:"text,omitempty"`
	// image block
	Source *anthropicImageSource `json:"source,omitempty"`
	// tool_use block
	ID    string                 `json:"id,omitempty"`
	Name  string                 `json:"name,omitempty"`
	Input map[string]interface{} `json:"input,omitempty"`
	// tool_result block
	ToolUseID string      `json:"tool_use_id,omitempty"`
	Content   interface{} `json:"content,omitempty"`
}

type anthropicImageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type,omitempty"`
	Data      string `json:"data,omitempty"`
	URL       string `json:"url,omitempty"`
}

type anthropicTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"input_schema,omitempty"`
}

type anthropicResponse struct {
	ID         string                  `json:"id"`
	Type       string                  `json:"type"`
	Role       string                  `json:"role"`
	Model      string                  `json:"model"`
	Content    []anthropicContentBlock `json:"content"`
	StopReason string                  `json:"stop_reason"`
	Usage      anthropicUsage          `json:"usage"`
}

type anthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

func (o *OpenAIChat) RequestConvert(ctx eocontext.EoContext, extender map[string]interface{}) error {
	context_label2.SetBillingMode(ctx, context_label2.BillingModeImmediate)
	context_label2.SetModelCompletionTag(ctx)
	httpContext, err := http_service.Assert(ctx)
	if err != nil {
		return err
	}
	body, err := httpContext.Proxy().Body().RawBody()
	if err != nil {
		return err
	}
	chatRequest := eosc.NewBase[openai.ChatCompletionRequest](extender)
	err = json.Unmarshal(body, chatRequest)
	if err != nil {
		return fmt.Errorf("unmarshal body error: %v, body: %s", err, string(body))
	}
	if chatRequest.Config.Model == "" {
		chatRequest.Config.Model = ai_convert.GetAIModel(ctx)
	}

	anthropicReq := convertOpenAIToAnthropic(chatRequest.Config)

	newBody, err := json.Marshal(anthropicReq)
	if err != nil {
		return fmt.Errorf("marshal anthropic request error: %v", err)
	}

	httpContext.Proxy().Header().SetHeader("anthropic-version", defaultVersion)
	if o.apikey != "" {
		httpContext.Proxy().Header().SetHeader("x-api-key", o.apikey)
	}
	httpContext.Proxy().Header().DelHeader("authorization")
	httpContext.Proxy().URI().SetPath(o.path)
	httpContext.Proxy().Body().SetRaw("application/json", newBody)

	if chatRequest.Config.Stream {
		httpContext.Proxy().AppendStreamBodyHandle(o.streamHandler)
	} else {
		context_label2.SetDisableStream(ctx, true)
	}

	if o.balanceHandler != nil {
		ctx.SetBalance(o.balanceHandler)
	}
	return nil
}

// convertOpenAIToAnthropic converts an OpenAI chat completion request into the
// Anthropic Messages API request format.
func convertOpenAIToAnthropic(req *openai.ChatCompletionRequest) *anthropicRequest {
	out := &anthropicRequest{
		Model:  req.Model,
		Stream: req.Stream,
	}
	if req.MaxTokens > 0 {
		out.MaxTokens = req.MaxTokens
	} else {
		out.MaxTokens = defaultMaxTokens
	}
	if req.Temperature != 0 {
		t := req.Temperature
		out.Temperature = &t
	}
	if req.TopP != 0 {
		p := req.TopP
		out.TopP = &p
	}
	if len(req.Stop) > 0 {
		out.StopSequences = req.Stop
	}

	systemBuilder := strings.Builder{}
	for _, msg := range req.Messages {
		switch msg.Role {
		case "system", "developer":
			if systemBuilder.Len() > 0 {
				systemBuilder.WriteString("\n")
			}
			systemBuilder.WriteString(msg.Content)
			continue
		case "assistant":
			blocks := make([]anthropicContentBlock, 0)
			if msg.Content != "" {
				blocks = append(blocks, anthropicContentBlock{Type: "text", Text: msg.Content})
			}
			for _, tc := range msg.ToolCalls {
				var input map[string]interface{}
				if tc.Function.Arguments != "" {
					_ = json.Unmarshal([]byte(tc.Function.Arguments), &input)
				}
				blocks = append(blocks, anthropicContentBlock{
					Type:  "tool_use",
					ID:    tc.ID,
					Name:  tc.Function.Name,
					Input: input,
				})
			}
			if len(blocks) > 0 {
				out.Messages = append(out.Messages, anthropicMessage{Role: "assistant", Content: blocks})
			}
		case "tool":
			// tool result must be sent as a user message in Anthropic format.
			out.Messages = append(out.Messages, anthropicMessage{
				Role: "user",
				Content: []anthropicContentBlock{{
					Type:      "tool_result",
					ToolUseID: msg.ToolCallID,
					Content:   msg.Content,
				}},
			})
		default: // user
			out.Messages = append(out.Messages, anthropicMessage{
				Role:    "user",
				Content: convertUserContent(msg),
			})
		}
	}
	out.System = systemBuilder.String()

	// tools
	for _, tool := range req.Tools {
		if tool.Type != openai.ToolTypeFunction || tool.Function == nil {
			continue
		}
		at := anthropicTool{
			Name:        tool.Function.Name,
			Description: tool.Function.Description,
		}
		if tool.Function.Parameters != nil {
			if b, err := json.Marshal(tool.Function.Parameters); err == nil {
				var schema map[string]interface{}
				if err := json.Unmarshal(b, &schema); err == nil {
					at.InputSchema = schema
				}
			}
		}
		out.Tools = append(out.Tools, at)
	}

	// tool_choice
	if req.ToolChoice != nil {
		out.ToolChoice = convertToolChoice(req.ToolChoice)
	}

	return out
}

func convertUserContent(msg openai.ChatCompletionMessage) []anthropicContentBlock {
	if len(msg.MultiContent) == 0 {
		return []anthropicContentBlock{{Type: "text", Text: msg.Content}}
	}
	blocks := make([]anthropicContentBlock, 0, len(msg.MultiContent))
	for _, part := range msg.MultiContent {
		switch part.Type {
		case openai.ChatMessagePartTypeText:
			blocks = append(blocks, anthropicContentBlock{Type: "text", Text: part.Text})
		case openai.ChatMessagePartTypeImageURL:
			if part.ImageURL == nil {
				continue
			}
			u := part.ImageURL.URL
			if strings.HasPrefix(u, "data:") {
				// data:<mediaType>;base64,<data>
				mediaType := "image/jpeg"
				data := u
				if commaIdx := strings.Index(u, ","); commaIdx != -1 {
					meta := u[len("data:"):commaIdx]
					data = u[commaIdx+1:]
					if semiIdx := strings.Index(meta, ";"); semiIdx != -1 {
						mediaType = meta[:semiIdx]
					} else {
						mediaType = meta
					}
				}
				blocks = append(blocks, anthropicContentBlock{
					Type: "image",
					Source: &anthropicImageSource{
						Type:      "base64",
						MediaType: mediaType,
						Data:      data,
					},
				})
			} else if u != "" {
				blocks = append(blocks, anthropicContentBlock{
					Type: "image",
					Source: &anthropicImageSource{
						Type: "url",
						URL:  u,
					},
				})
			}
		}
	}
	return blocks
}

func convertToolChoice(toolChoice interface{}) interface{} {
	b, err := json.Marshal(toolChoice)
	if err != nil {
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		switch s {
		case "none":
			return map[string]interface{}{"type": "none"}
		case "auto":
			return map[string]interface{}{"type": "auto"}
		case "required":
			return map[string]interface{}{"type": "any"}
		}
		return nil
	}
	var obj struct {
		Type     string `json:"type"`
		Function struct {
			Name string `json:"name"`
		} `json:"function"`
	}
	if err := json.Unmarshal(b, &obj); err == nil && obj.Function.Name != "" {
		return map[string]interface{}{"type": "tool", "name": obj.Function.Name}
	}
	return nil
}

func (o *OpenAIChat) ResponseConvert(ctx eocontext.EoContext) error {
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
		o.convertErrorResponse(httpContext, body)
		return nil
	}

	var resp anthropicResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		ai_convert.SetAIProviderStatuses(httpContext, ai_convert.StatusInvalid)
		log.Errorf("unmarshal body error: %v, body: %s", err, string(body))
		return err
	}

	ai_convert.SetAIModelInputToken(httpContext, resp.Usage.InputTokens)
	ai_convert.SetAIModelOutputToken(httpContext, resp.Usage.OutputTokens)
	ai_convert.SetAIModelTotalToken(httpContext, resp.Usage.InputTokens+resp.Usage.OutputTokens)
	ai_convert.SetAIStatusNormal(ctx)
	ai_convert.SetAIProviderStatuses(httpContext, ai_convert.GetAIStatus(ctx))

	openaiResp := convertAnthropicToOpenAI(ctx.RequestId(), ai_convert.GetAIModel(ctx), &resp)
	newBody, err := json.Marshal(openaiResp)
	if err != nil {
		return err
	}
	httpContext.Response().SetHeader("content-encoding", "utf-8")
	httpContext.Response().SetBody(newBody)
	return nil
}

// convertAnthropicToOpenAI converts an Anthropic Messages response into the
// OpenAI chat completion response format.
func convertAnthropicToOpenAI(requestID string, model string, resp *anthropicResponse) *openai.ChatCompletionResponse {
	out := &openai.ChatCompletionResponse{
		ID:      "chatcmpl-" + requestID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   model,
		Usage: openai.Usage{
			PromptTokens:     resp.Usage.InputTokens,
			CompletionTokens: resp.Usage.OutputTokens,
			TotalTokens:      resp.Usage.InputTokens + resp.Usage.OutputTokens,
		},
	}

	choice := openai.ChatCompletionChoice{
		Index: 0,
		Message: openai.ChatCompletionMessage{
			Role: "assistant",
		},
	}
	contentBuilder := strings.Builder{}
	var toolCalls []openai.ToolCall
	for _, block := range resp.Content {
		switch block.Type {
		case "text":
			contentBuilder.WriteString(block.Text)
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
		}
	}
	choice.Message.Content = contentBuilder.String()
	if len(toolCalls) > 0 {
		choice.Message.ToolCalls = toolCalls
	}
	choice.FinishReason = mapStopReason(resp.StopReason, len(toolCalls) > 0)

	out.Choices = []openai.ChatCompletionChoice{choice}
	return out
}

func mapStopReason(reason string, hasToolCalls bool) openai.FinishReason {
	switch reason {
	case "end_turn", "stop_sequence":
		return openai.FinishReasonStop
	case "max_tokens":
		return openai.FinishReasonLength
	case "tool_use":
		return openai.FinishReasonToolCalls
	default:
		if hasToolCalls {
			return openai.FinishReasonToolCalls
		}
		return openai.FinishReasonStop
	}
}

// convertErrorResponse converts an Anthropic error body into the OpenAI error format.
func (o *OpenAIChat) convertErrorResponse(httpContext http_service.IHttpContext, body []byte) {
	var anthropicErr struct {
		Type  string `json:"type"`
		Error struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &anthropicErr); err != nil || anthropicErr.Error.Message == "" {
		return
	}
	openaiErr := map[string]interface{}{
		"error": map[string]interface{}{
			"message": anthropicErr.Error.Message,
			"type":    anthropicErr.Error.Type,
			"code":    httpContext.Response().StatusCode(),
		},
	}
	newBody, err := json.Marshal(openaiErr)
	if err != nil {
		return
	}
	httpContext.Response().SetBody(newBody)
}

// streamHandler converts an Anthropic SSE stream into an OpenAI-compatible SSE stream.
func (o *OpenAIChat) streamHandler(ctx http_service.IHttpContext, p []byte) ([]byte, error) {
	var sseBuffer bytes.Buffer
	requestID := "chatcmpl-" + ctx.RequestId()
	model := ai_convert.GetAIModel(ctx)

	// The framework may hand us partial SSE events; buffer the incomplete tail.
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

	var eventType string
	// track tool_use blocks by content index for streaming arguments.
	inputTokens := ctx.GetLabel(labelStreamInputTokens)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "event:") {
			eventType = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			continue
		}
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		dataStr := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if dataStr == "" {
			continue
		}

		switch eventType {
		case "message_start":
			var evt struct {
				Message struct {
					Usage anthropicUsage `json:"usage"`
				} `json:"message"`
			}
			if err := json.Unmarshal([]byte(dataStr), &evt); err == nil {
				inputTokens = fmt.Sprintf("%d", evt.Message.Usage.InputTokens)
				ctx.SetLabel(labelStreamInputTokens, inputTokens)
				ai_convert.SetAIModelInputToken(ctx, evt.Message.Usage.InputTokens)
			}
		case "content_block_start":
			var evt struct {
				Index        int                   `json:"index"`
				ContentBlock anthropicContentBlock `json:"content_block"`
			}
			if err := json.Unmarshal([]byte(dataStr), &evt); err == nil && evt.ContentBlock.Type == "tool_use" {
				idx := evt.Index
				chunk := openai.ChatCompletionStreamResponse{
					ID:      requestID,
					Object:  "chat.completion.chunk",
					Created: time.Now().Unix(),
					Model:   model,
					Choices: []openai.ChatCompletionStreamChoice{{
						Index: 0,
						Delta: openai.ChatCompletionStreamChoiceDelta{
							Role: "assistant",
							ToolCalls: []openai.ToolCall{{
								Index: &idx,
								ID:    evt.ContentBlock.ID,
								Type:  openai.ToolTypeFunction,
								Function: openai.FunctionCall{
									Name:      evt.ContentBlock.Name,
									Arguments: "",
								},
							}},
						},
					}},
				}
				writeSSE(&sseBuffer, chunk)
			}
		case "content_block_delta":
			var evt struct {
				Index int `json:"index"`
				Delta struct {
					Type        string `json:"type"`
					Text        string `json:"text"`
					PartialJSON string `json:"partial_json"`
				} `json:"delta"`
			}
			if err := json.Unmarshal([]byte(dataStr), &evt); err != nil {
				continue
			}
			switch evt.Delta.Type {
			case "text_delta":
				if evt.Delta.Text == "" {
					continue
				}
				chunk := openai.ChatCompletionStreamResponse{
					ID:      requestID,
					Object:  "chat.completion.chunk",
					Created: time.Now().Unix(),
					Model:   model,
					Choices: []openai.ChatCompletionStreamChoice{{
						Index: 0,
						Delta: openai.ChatCompletionStreamChoiceDelta{
							Role:    "assistant",
							Content: evt.Delta.Text,
						},
					}},
				}
				writeSSE(&sseBuffer, chunk)
			case "input_json_delta":
				if evt.Delta.PartialJSON == "" {
					continue
				}
				idx := evt.Index
				chunk := openai.ChatCompletionStreamResponse{
					ID:      requestID,
					Object:  "chat.completion.chunk",
					Created: time.Now().Unix(),
					Model:   model,
					Choices: []openai.ChatCompletionStreamChoice{{
						Index: 0,
						Delta: openai.ChatCompletionStreamChoiceDelta{
							Role: "assistant",
							ToolCalls: []openai.ToolCall{{
								Index: &idx,
								Function: openai.FunctionCall{
									Arguments: evt.Delta.PartialJSON,
								},
							}},
						},
					}},
				}
				writeSSE(&sseBuffer, chunk)
			}
		case "message_delta":
			var evt struct {
				Delta struct {
					StopReason string `json:"stop_reason"`
				} `json:"delta"`
				Usage anthropicUsage `json:"usage"`
			}
			if err := json.Unmarshal([]byte(dataStr), &evt); err != nil {
				continue
			}
			input := 0
			fmt.Sscanf(inputTokens, "%d", &input)
			output := evt.Usage.OutputTokens
			ai_convert.SetAIModelInputToken(ctx, input)
			ai_convert.SetAIModelOutputToken(ctx, output)
			ai_convert.SetAIModelTotalToken(ctx, input+output)

			chunk := openai.ChatCompletionStreamResponse{
				ID:      requestID,
				Object:  "chat.completion.chunk",
				Created: time.Now().Unix(),
				Model:   model,
				Choices: []openai.ChatCompletionStreamChoice{{
					Index:        0,
					Delta:        openai.ChatCompletionStreamChoiceDelta{},
					FinishReason: mapStopReason(evt.Delta.StopReason, false),
				}},
				Usage: &openai.Usage{
					PromptTokens:     input,
					CompletionTokens: output,
					TotalTokens:      input + output,
				},
			}
			writeSSE(&sseBuffer, chunk)
		case "message_stop":
			sseBuffer.WriteString("data: [DONE]\n\n")
		}
		eventType = ""
	}

	return sseBuffer.Bytes(), nil
}

func writeSSE(buf *bytes.Buffer, chunk openai.ChatCompletionStreamResponse) {
	content, _ := json.Marshal(chunk)
	buf.WriteString(fmt.Sprintf("data: %s\n\n", string(content)))
}

const (
	labelStreamRemain      = "anthropic_stream_remain"
	labelStreamInputTokens = "anthropic_stream_input_tokens"
)
