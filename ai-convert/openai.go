package ai_convert

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	context_label "github.com/eolinker/apinto/utils/context-label"
	"net/url"
	"strings"
	"time"

	"github.com/eolinker/eosc/log"

	http_service "github.com/eolinker/eosc/eocontext/http-context"

	"github.com/eolinker/apinto/encoder"
	"github.com/eolinker/eosc"
	eoscContext "github.com/eolinker/eosc/eocontext"
	openai "github.com/sashabaranov/go-openai"
)

const (
	OpenAIChatCompletePath = "/chat/completions"
)

type CheckError func(ctx http_service.IHttpContext, body []byte) bool

type OpenAIChat struct {
	provider       string
	apikey         string
	path           string
	modelType      ModelType
	checkErr       CheckError
	errorCallback  func(ctx http_service.IHttpContext, body []byte)
	balanceHandler eoscContext.BalanceHandler
}

func NewOpenAIChat(provider string, apikey string, baseUrl string, modelType ModelType, timeout time.Duration, checkErr CheckError, errorCallback func(ctx http_service.IHttpContext, body []byte)) (IConverterDriver, error) {
	c := &OpenAIChat{
		provider:      provider,
		apikey:        apikey,
		modelType:     modelType,
		checkErr:      checkErr,
		errorCallback: errorCallback,
	}
	if baseUrl != "" {
		balanceHandler, err := NewBalanceHandler(apikey, baseUrl, timeout)
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
		c.path = fmt.Sprintf("%s%s", strings.TrimSuffix(u.Path, "/"), OpenAIChatCompletePath)
	} else {
		c.path = fmt.Sprintf("/v1%s", OpenAIChatCompletePath)
	}
	return c, nil
}

func (o *OpenAIChat) Provider() string {
	return o.provider
}

func (o *OpenAIChat) ModelType() ModelType {
	return o.modelType
}

func (o *OpenAIChat) RequestConvert(ctx eoscContext.EoContext, extender map[string]interface{}) error {
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
	err = json.Unmarshal(body, chatRequest)
	if err != nil {
		return fmt.Errorf("unmarshal body error: %v, body: %s", err, string(body))
	}
	if chatRequest.Config.Model == "" {
		chatRequest.Config.Model = GetAIModel(ctx)
	}
	if chatRequest.Config.Stream {
		chatRequest.Config.StreamOptions = &openai.StreamOptions{
			IncludeUsage: true,
		}
	}
	totalMessageBuilder := strings.Builder{}
	for _, msg := range chatRequest.Config.Messages {
		totalMessageBuilder.WriteString(msg.Content)
	}
	promptToken := getTokens(totalMessageBuilder.String(), chatRequest.Config.Model)
	SetAIModelInputToken(httpContext, promptToken)
	if o.apikey != "" {
		httpContext.Proxy().Header().SetHeader("Authorization", "Bearer "+o.apikey)
	}
	httpContext.Proxy().URI().SetPath(o.path)
	body, _ = json.Marshal(chatRequest)
	httpContext.Proxy().Body().SetRaw("application/json", body)
	if o.balanceHandler != nil {
		ctx.SetBalance(o.balanceHandler)
	}
	httpContext.Proxy().SetStreamBodyParse(StreamBodyParse)
	httpContext.Proxy().AppendBodyFinish(o.bodyFinish)

	return nil
}

func ResponseConvert(ctx eoscContext.EoContext, checkErr CheckError, errorCallback func(ctx http_service.IHttpContext, body []byte)) error {
	httpContext, err := http_service.Assert(ctx)
	if err != nil {
		return err
	}
	body := httpContext.Response().GetBody()
	// Check the content encoding and convert to UTF-8 if necessary.
	encoding := httpContext.Response().Headers().Get("content-encoding")
	if encoding != "utf-8" && encoding != "" {
		body, err = encoder.ToUTF8(encoding, body)
		if err != nil {
			return err
		}
	}

	if (checkErr != nil && !checkErr(httpContext, body)) || httpContext.Response().StatusCode() != 200 {
		if errorCallback != nil {
			errorCallback(httpContext, body)
		}
		status := GetAIStatus(ctx)
		if status == "" {
			status = StatusInvalid
		}
		SetAIProviderStatuses(httpContext, status)
		return nil
	}

	var resp openai.ChatCompletionResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		SetAIProviderStatuses(httpContext, StatusInvalid)
		log.Errorf("unmarshal body error: %v, body: %s", err, string(body))
		return err
	}

	SetAIModelInputToken(httpContext, resp.Usage.PromptTokens)
	SetAIModelOutputToken(httpContext, resp.Usage.CompletionTokens)
	SetAIModelTotalToken(httpContext, resp.Usage.TotalTokens)
	SetAIStatusNormal(ctx)
	SetAIProviderStatuses(httpContext, GetAIStatus(ctx))
	httpContext.Response().SetHeader("content-encoding", "utf-8")
	httpContext.Response().SetBody(body)
	return nil
}

func (o *OpenAIChat) ResponseConvert(ctx eoscContext.EoContext) error {
	return ResponseConvert(ctx, o.checkErr, o.errorCallback)
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
		line = strings.TrimPrefix(line, "data:")
		if line == "" || strings.Trim(line, " ") == "[DONE]" {
			continue
		}
		builder.WriteString(line)
		builder.WriteString("\n")
	}
	return []byte(builder.String())
}

func (o *OpenAIChat) bodyFinish(ctx http_service.IHttpContext) {
	body := ctx.Response().GetBody()
	defer func() {
		SetAIProviderStatuses(ctx, GetAIStatus(ctx))
	}()
	if o.checkErr != nil && !o.checkErr(ctx, body) {
		o.errorCallback(ctx, body)
		return
	}
	encoding := ctx.Response().Headers().Get("content-encoding")
	if encoding != "utf-8" && encoding != "" {
		tmp, err := encoder.ToUTF8(encoding, body)
		if err != nil {
			log.Errorf("convert to utf-8 error: %v, body: %s", err, string(body))
			return
		}
		body = tmp
	}
	if context_label.IsStreamRunning(ctx) {
		usage, err := calculateStreamOutputToken(body, GetAIModel(ctx))
		if err != nil {
			log.Errorf("calculate stream output error: %v, body: %s", err, string(body))
			return
		}
		input := usage.Input
		if input == 0 {
			input = GetAIModelInputToken(ctx)
		}
		output := usage.Output
		total := usage.Total
		if total == 0 {
			total = input + output
		}
		SetAIModelInputToken(ctx, input)
		SetAIModelOutputToken(ctx, output)
		SetAIModelTotalToken(ctx, total)
	} else {
		var resp openai.ChatCompletionResponse
		err := json.Unmarshal(body, &resp)
		if err != nil {
			log.Errorf("unmarshal body error: %v, body: %s", err, string(body))
			return
		}
		SetAIModelInputToken(ctx, resp.Usage.PromptTokens)
		SetAIModelOutputToken(ctx, resp.Usage.CompletionTokens)
		SetAIModelTotalToken(ctx, resp.Usage.TotalTokens)
	}
	SetAIStatusNormal(ctx)
}

type TokenUsage struct {
	Input  int
	Output int
	Total  int
}

func calculateStreamOutputToken(body []byte, model string) (*TokenUsage, error) {
	builder := strings.Builder{}
	scanner := bufio.NewScanner(bytes.NewReader(body))
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimPrefix(line, "data:")
		if line == "" || strings.Trim(line, " ") == "[DONE]" {
			continue
		}
		var resp openai.ChatCompletionStreamResponse
		err := json.Unmarshal([]byte(line), &resp)
		if err != nil {
			log.Errorf("unmarshal stream body error: %v, body: %s", err, string(body))
			continue
		}
		if len(resp.Choices) > 0 {
			builder.WriteString(resp.Choices[0].Delta.Content)

		}
		if resp.Usage != nil {
			return &TokenUsage{resp.Usage.PromptTokens, resp.Usage.CompletionTokens, resp.Usage.TotalTokens}, nil
		}
	}
	return &TokenUsage{0, getTokens(builder.String(), model), 0}, nil
}
