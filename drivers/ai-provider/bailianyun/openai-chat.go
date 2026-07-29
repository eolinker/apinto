package bailianyun

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

var (
	openaiChatPath = "/services/aigc/text-generation/generation"
)

const bailianStreamRemainLabel = "bailian-openai-stream-remain"

func init() {
	driverCreate.Set(ai_convert.ModelTypeOpenAIChat, func(mt ai_convert.ModelType, c *Config) (ai_convert.IConverterDriver, error) {
		return NewOpenAIChat(provider, c.APIKey, c.BaseUrl, mt, 10*time.Minute)
	})
}

func NewOpenAIChat(provider string, apikey string, baseUrl string, modelType ai_convert.ModelType, timeout time.Duration) (ai_convert.IConverterDriver, error) {
	c := &OpenAIChat{
		provider:  provider,
		apikey:    apikey,
		modelType: modelType,
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
		c.path = "/api/v1" + openaiChatPath
	} else {
		c.path = fmt.Sprintf("%s%s", strings.TrimSuffix(u.Path, "/"), openaiChatPath)
	}
	return c, nil
}

type OpenAIChat struct {
	apikey         string
	provider       string
	path           string
	checkErr       ai_convert.CheckError
	errorCallback  func(ctx http_service.IHttpContext, body []byte)
	balanceHandler eocontext.BalanceHandler
	modelType      ai_convert.ModelType
}

func (c *OpenAIChat) Provider() string {
	return c.provider
}

func (c *OpenAIChat) ModelType() ai_convert.ModelType {
	return c.modelType
}

func (c *OpenAIChat) RequestConvert(ctx eocontext.EoContext, extender map[string]interface{}) error {
	context_label.SetBillingMode(ctx, context_label.BillingModeImmediate)
	httpContext, err := http_service.Assert(ctx)
	if err != nil {
		return err
	}

	body, err := httpContext.Proxy().Body().RawBody()
	if err != nil {
		return err
	}
	chatRequest := eosc.NewBase[openai.ChatCompletionRequest](extender)
	if err = json.Unmarshal(body, chatRequest); err != nil {
		return fmt.Errorf("unmarshal body error: %v, body: %s", err, string(body))
	}
	if chatRequest.Config.Model == "" {
		chatRequest.Config.Model = ai_convert.GetAIModel(ctx)
	}

	dashScopeReq := convertOpenAIToDashScope(chatRequest.Config, chatRequest.Append)
	newBody, err := json.Marshal(dashScopeReq)
	if err != nil {
		return fmt.Errorf("marshal dashscope request error: %v", err)
	}

	if c.apikey != "" {
		httpContext.Proxy().Header().SetHeader("Authorization", "Bearer "+c.apikey)
	}
	httpContext.Proxy().URI().SetPath(c.path)
	context_label.SetModelCompletionTag(ctx)
	httpContext.Proxy().Body().SetRaw("application/json", newBody)
	if chatRequest.Config.Stream {
		httpContext.Proxy().Header().SetHeader("X-DashScope-SSE", "enable")
		httpContext.Proxy().AppendStreamBodyHandle(c.streamHandler)
		httpContext.Proxy().SetStreamBodyParse(StreamBodyParse)
	} else {
		context_label.SetDisableStream(ctx, true)
	}
	if c.balanceHandler != nil {
		ctx.SetBalance(c.balanceHandler)
	}
	return nil
}

func (c *OpenAIChat) ResponseConvert(ctx eocontext.EoContext) error {
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
		c.convertErrorResponse(httpContext, body)
		return nil
	}
	newBody, err := convertDashScopeToOpenAI(httpContext, body)
	if err != nil {
		log.Errorf("convert dashscope response to openai error: %v, body: %s", err, string(body))
		return err
	}
	httpContext.Response().SetHeader("content-encoding", "utf-8")
	httpContext.Response().SetBody(newBody)
	return nil
}

type dashScopeRequest struct {
	Model      string                 `json:"model"`
	Input      dashScopeInput         `json:"input"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

type dashScopeInput struct {
	Messages []dashScopeMessage `json:"messages"`
}

type dashScopeMessage struct {
	Role       string            `json:"role"`
	Content    interface{}       `json:"content,omitempty"`
	Name       string            `json:"name,omitempty"`
	ToolCalls  []openai.ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string            `json:"tool_call_id,omitempty"`
}

type dashScopeResponse struct {
	Output    dashScopeOutput `json:"output"`
	Usage     dashScopeUsage  `json:"usage"`
	RequestID string          `json:"request_id"`
	Code      string          `json:"code"`
	Message   string          `json:"message"`
}

type dashScopeOutput struct {
	Text         string            `json:"text"`
	FinishReason string            `json:"finish_reason"`
	Choices      []dashScopeChoice `json:"choices"`
}

type dashScopeChoice struct {
	FinishReason string           `json:"finish_reason"`
	Message      dashScopeMessage `json:"message"`
}

type dashScopeUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

func convertOpenAIToDashScope(req *openai.ChatCompletionRequest, appendParams map[string]interface{}) *dashScopeRequest {
	parameters := make(map[string]interface{})
	for k, v := range appendParams {
		parameters[k] = v
	}
	parameters["result_format"] = "message"
	if req.MaxTokens > 0 {
		parameters["max_tokens"] = req.MaxTokens
	} else if req.MaxCompletionTokens > 0 {
		parameters["max_tokens"] = req.MaxCompletionTokens
	}
	if req.Temperature != 0 {
		parameters["temperature"] = req.Temperature
	}
	if req.TopP != 0 {
		parameters["top_p"] = req.TopP
	}
	if len(req.Stop) > 0 {
		parameters["stop"] = req.Stop
	}
	if req.Seed != nil {
		parameters["seed"] = req.Seed
	}
	if req.PresencePenalty != 0 {
		parameters["presence_penalty"] = req.PresencePenalty
	}
	if req.FrequencyPenalty != 0 {
		parameters["frequency_penalty"] = req.FrequencyPenalty
	}
	if req.ResponseFormat != nil {
		parameters["response_format"] = req.ResponseFormat
	}
	if len(req.Tools) > 0 {
		parameters["tools"] = req.Tools
	}
	if req.ToolChoice != nil {
		parameters["tool_choice"] = req.ToolChoice
	}
	if req.ParallelToolCalls != nil {
		parameters["parallel_tool_calls"] = req.ParallelToolCalls
	}
	if req.Stream {
		parameters["incremental_output"] = true
	}

	messages := make([]dashScopeMessage, 0, len(req.Messages))
	for _, msg := range req.Messages {
		messages = append(messages, convertOpenAIMessageToDashScope(msg))
	}
	return &dashScopeRequest{
		Model:      req.Model,
		Input:      dashScopeInput{Messages: messages},
		Parameters: parameters,
	}
}

func convertOpenAIMessageToDashScope(msg openai.ChatCompletionMessage) dashScopeMessage {
	out := dashScopeMessage{
		Role:       msg.Role,
		Name:       msg.Name,
		ToolCalls:  msg.ToolCalls,
		ToolCallID: msg.ToolCallID,
	}
	if len(msg.MultiContent) == 0 {
		out.Content = msg.Content
		return out
	}
	parts := make([]map[string]interface{}, 0, len(msg.MultiContent))
	for _, part := range msg.MultiContent {
		switch part.Type {
		case openai.ChatMessagePartTypeText:
			parts = append(parts, map[string]interface{}{"text": part.Text})
		case openai.ChatMessagePartTypeImageURL:
			if part.ImageURL != nil && part.ImageURL.URL != "" {
				parts = append(parts, map[string]interface{}{"image": part.ImageURL.URL})
			}
		}
	}
	out.Content = parts
	return out
}

func convertDashScopeToOpenAI(ctx http_service.IHttpContext, body []byte) ([]byte, error) {
	var resp dashScopeResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	requestID := resp.RequestID
	if requestID == "" {
		requestID = ctx.RequestId()
	}
	model := ai_convert.GetAIModel(ctx)
	openaiResp := openai.ChatCompletionResponse{
		ID:      "chatcmpl-" + requestID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   model,
		Usage: openai.Usage{
			PromptTokens:     resp.Usage.InputTokens,
			CompletionTokens: resp.Usage.OutputTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}
	ai_convert.SetAIModelInputToken(ctx, resp.Usage.InputTokens)
	ai_convert.SetAIModelOutputToken(ctx, resp.Usage.OutputTokens)
	ai_convert.SetAIModelTotalToken(ctx, resp.Usage.TotalTokens)
	ai_convert.SetAIStatusNormal(ctx)
	ai_convert.SetAIProviderStatuses(ctx, ai_convert.GetAIStatus(ctx))

	if len(resp.Output.Choices) == 0 {
		openaiResp.Choices = []openai.ChatCompletionChoice{{
			Index: 0,
			Message: openai.ChatCompletionMessage{
				Role:    "assistant",
				Content: resp.Output.Text,
			},
			FinishReason: mapDashScopeFinishReason(resp.Output.FinishReason, false),
		}}
	} else {
		for index, choice := range resp.Output.Choices {
			message := openai.ChatCompletionMessage{Role: choice.Message.Role}
			if message.Role == "" {
				message.Role = "assistant"
			}
			message.Content = dashScopeContentToString(choice.Message.Content)
			message.ToolCalls = choice.Message.ToolCalls
			openaiResp.Choices = append(openaiResp.Choices, openai.ChatCompletionChoice{
				Index:        index,
				Message:      message,
				FinishReason: mapDashScopeFinishReason(choice.FinishReason, len(message.ToolCalls) > 0),
			})
		}
	}
	return json.Marshal(openaiResp)
}

func (c *OpenAIChat) streamHandler(ctx http_service.IHttpContext, p []byte) ([]byte, error) {
	var sseBuffer bytes.Buffer
	data := ctx.GetLabel(bailianStreamRemainLabel) + string(p)
	lastNL := strings.LastIndexByte(data, '\n')
	if lastNL < 0 {
		ctx.SetLabel(bailianStreamRemainLabel, data)
		return []byte{}, nil
	}
	ctx.SetLabel(bailianStreamRemainLabel, data[lastNL+1:])
	complete := data[:lastNL+1]
	scanner := bufio.NewScanner(strings.NewReader(complete))
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "" {
			continue
		}
		if payload == "[DONE]" {
			sseBuffer.WriteString("data: [DONE]\n\n")
			continue
		}
		var resp dashScopeResponse
		if err := json.Unmarshal([]byte(payload), &resp); err != nil {
			log.Errorf("unmarshal dashscope stream body error: %v, body: %s", err, payload)
			continue
		}
		chunks := convertDashScopeStreamToOpenAI(ctx, &resp)
		for _, chunk := range chunks {
			content, _ := json.Marshal(chunk)
			sseBuffer.WriteString(fmt.Sprintf("data: %s\n\n", content))
		}
	}
	return sseBuffer.Bytes(), nil
}

func convertDashScopeStreamToOpenAI(ctx http_service.IHttpContext, resp *dashScopeResponse) []openai.ChatCompletionStreamResponse {
	model := ai_convert.GetAIModel(ctx)
	requestID := resp.RequestID
	if requestID == "" {
		requestID = ctx.RequestId()
	}
	chunks := make([]openai.ChatCompletionStreamResponse, 0)
	for index, choice := range resp.Output.Choices {
		toolCalls := choice.Message.ToolCalls
		chunk := openai.ChatCompletionStreamResponse{
			ID:      "chatcmpl-" + requestID,
			Object:  "chat.completion.chunk",
			Created: time.Now().Unix(),
			Model:   model,
			Choices: []openai.ChatCompletionStreamChoice{{
				Index: index,
				Delta: openai.ChatCompletionStreamChoiceDelta{
					Role:      choice.Message.Role,
					Content:   dashScopeContentToString(choice.Message.Content),
					ToolCalls: toolCalls,
				},
				FinishReason: mapDashScopeFinishReason(choice.FinishReason, len(toolCalls) > 0),
			}},
		}
		if chunk.Choices[0].Delta.Role == "" {
			chunk.Choices[0].Delta.Role = "assistant"
		}
		chunks = append(chunks, chunk)
	}
	if resp.Usage.TotalTokens > 0 || resp.Usage.InputTokens > 0 || resp.Usage.OutputTokens > 0 {
		usage := openai.Usage{
			PromptTokens:     resp.Usage.InputTokens,
			CompletionTokens: resp.Usage.OutputTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		}
		ai_convert.SetAIModelInputToken(ctx, resp.Usage.InputTokens)
		ai_convert.SetAIModelOutputToken(ctx, resp.Usage.OutputTokens)
		ai_convert.SetAIModelTotalToken(ctx, resp.Usage.TotalTokens)
		chunks = append(chunks, openai.ChatCompletionStreamResponse{
			ID:      "chatcmpl-" + requestID,
			Object:  "chat.completion.chunk",
			Created: time.Now().Unix(),
			Model:   model,
			Choices: []openai.ChatCompletionStreamChoice{},
			Usage:   &usage,
		})
	}
	return chunks
}

func dashScopeContentToString(content interface{}) string {
	switch v := content.(type) {
	case string:
		return v
	case []interface{}:
		builder := strings.Builder{}
		for _, item := range v {
			m, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			if text, ok := m["text"].(string); ok {
				builder.WriteString(text)
			}
		}
		return builder.String()
	case nil:
		return ""
	default:
		b, _ := json.Marshal(v)
		return string(b)
	}
}

func mapDashScopeFinishReason(reason string, hasToolCalls bool) openai.FinishReason {
	if hasToolCalls {
		return openai.FinishReasonToolCalls
	}
	switch reason {
	case "stop", "null", "":
		return openai.FinishReasonStop
	case "length", "max_tokens":
		return openai.FinishReasonLength
	case "tool_calls":
		return openai.FinishReasonToolCalls
	case "content_filter":
		return openai.FinishReasonContentFilter
	default:
		return openai.FinishReasonStop
	}
}

func (c *OpenAIChat) convertErrorResponse(httpContext http_service.IHttpContext, body []byte) {
	var dashScopeErr dashScopeResponse
	if err := json.Unmarshal(body, &dashScopeErr); err != nil || dashScopeErr.Message == "" {
		return
	}
	openaiErr := map[string]interface{}{
		"error": map[string]interface{}{
			"message": dashScopeErr.Message,
			"type":    dashScopeErr.Code,
			"code":    dashScopeErr.Code,
		},
	}
	newBody, err := json.Marshal(openaiErr)
	if err != nil {
		return
	}
	httpContext.Response().SetBody(newBody)
}

func StreamBodyParse(ctx http_service.IHttpContext, body []byte) []byte {
	encoding := ctx.Response().Headers().Get("content-encoding")
	target := body
	if encoding != "utf-8" && encoding != "" {
		tmp, err := encoder.ToUTF8(encoding, body)
		if err != nil {
			log.Errorf("convert to utf-8 error: %v, body: %s", err, string(body))
			return body
		}
		target = tmp
	}
	builder := strings.Builder{}
	scanner := bufio.NewScanner(bytes.NewReader(target))
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 1 {
			continue
		}
		if strings.HasPrefix(line, "data:") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if line[0] == '{' && line[len(line)-1] == '}' {
				builder.WriteString(line)
				builder.WriteString("\n")
			}
		}
	}
	return []byte(builder.String())
}
