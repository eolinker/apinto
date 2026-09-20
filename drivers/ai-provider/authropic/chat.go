package anthropic

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	context_label "github.com/eolinker/apinto/common/context-label"
	"github.com/eolinker/eosc"
	"net/url"
	"strings"
	"time"

	ai_convert "github.com/eolinker/apinto/ai-convert"
	eoscContext "github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

func init() {
	driverCreate.Set(ai_convert.ModelTypeChat, func(mt ai_convert.ModelType, c *Config) (ai_convert.IConverterDriver, error) {
		return NewChat(provider, c.APIKey, c.Base, mt, 10*time.Minute)
	})
	driverCreate.Set(ai_convert.ModelTypeAnthropicChat, func(mt ai_convert.ModelType, c *Config) (ai_convert.IConverterDriver, error) {
		return NewChat(provider, c.APIKey, c.Base, mt, 10*time.Minute)
	})
}

const (
	chatPath = "/messages"
)

var _ ai_convert.IConverterDriver = (*Chat)(nil)

func NewChat(provider string, apikey string, baseUrl string, modelType ai_convert.ModelType, timeout time.Duration) (ai_convert.IConverterDriver, error) {
	c := &Chat{
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

// Chat 原生Chat模式
type Chat struct {
	provider       string
	apikey         string
	path           string
	modelType      ai_convert.ModelType
	balanceHandler eoscContext.BalanceHandler
}

func (c *Chat) Provider() string {
	return c.provider
}

func (c *Chat) ModelType() ai_convert.ModelType {
	return c.modelType
}

func (c *Chat) RequestConvert(ctx eoscContext.EoContext, extender map[string]interface{}) error {
	httpContext, err := http_service.Assert(ctx)
	context_label.SetModelCompletionTag(ctx)
	if err != nil {
		return err
	}
	body, err := httpContext.Proxy().Body().RawBody()
	if err != nil {
		return err
	}
	type ReqBody struct {
		Stream bool `json:"stream"`
	}
	chatRequest := eosc.NewBase[ReqBody](extender)
	err = json.Unmarshal(body, chatRequest)
	if err != nil {
		return fmt.Errorf("unmarshal body error: %v, body: %s", err, string(body))
	}
	if !chatRequest.Config.Stream {
		context_label.SetDisableStream(ctx, true)
	}

	//SetAIModelInputToken(httpContext, promptToken)
	httpContext.Proxy().Header().SetHeader("anthropic-version", defaultVersion)
	httpContext.Proxy().Header().SetHeader("x-api-key", c.apikey)
	httpContext.Proxy().Header().DelHeader("authorization")
	httpContext.Proxy().URI().SetPath(c.path)
	//body, _ = json.Marshal(chatRequest)
	//httpContext.Proxy().Body().SetRaw("application/json", body)
	if c.balanceHandler != nil {
		ctx.SetBalance(c.balanceHandler)
	}
	//httpContext.Proxy().AppendBodyFinish(c.bodyFinish)

	return nil
}

func (c *Chat) ResponseConvert(ctx eoscContext.EoContext) error {
	return nil
}

func calculateAnthropicStreamUsage(body []byte) (input, output int) {
	scanner := bufio.NewScanner(bytes.NewReader(body))
	var eventType string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "event: ") {
			eventType = strings.TrimPrefix(line, "event: ")
			continue
		}
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		switch eventType {
		case "message_start":
			var msg struct {
				Message struct {
					Usage struct {
						InputTokens int `json:"input_tokens"`
					} `json:"usage"`
				} `json:"message"`
			}
			if err := json.Unmarshal([]byte(data), &msg); err == nil {
				input = msg.Message.Usage.InputTokens
			}
		case "message_delta":
			var delta struct {
				Usage struct {
					OutputTokens int `json:"output_tokens"`
				} `json:"usage"`
			}
			if err := json.Unmarshal([]byte(data), &delta); err == nil {
				output += delta.Usage.OutputTokens
			}
		}
		eventType = ""
	}
	return
}

//func (c *Chat) bodyFinish(ctx http_service.IHttpContext) {
//	body := ctx.Response().GetBody()
//	defer func() {
//		ai_convert.SetAIProviderStatuses(ctx, ai_convert.GetAIStatus(ctx))
//	}()
//
//	encoding := ctx.Response().Headers().Get("content-encoding")
//	if encoding != "utf-8" && encoding != "" {
//		tmp, err := encoder.ToUTF8(encoding, body)
//		if err != nil {
//			log.Errorf("convert to utf-8 error: %v, body: %s", err, string(body))
//			return
//		}
//		body = tmp
//	}
//	if context_label.IsStreamRunning(ctx) {
//		input, output := calculateAnthropicStreamUsage(body)
//
//		if input == 0 {
//			input = ai_convert.GetAIModelInputToken(ctx)
//		}
//		total := input + output
//		ai_convert.SetAIModelInputToken(ctx, input)
//		ai_convert.SetAIModelOutputToken(ctx, output)
//		ai_convert.SetAIModelTotalToken(ctx, total)
//	} else {
//		var resp openai.ChatCompletionResponse
//		err := json.Unmarshal(body, &resp)
//		if err != nil {
//			log.Errorf("unmarshal body error: %v, body: %s", err, string(body))
//			return
//		}
//		ai_convert.SetAIModelInputToken(ctx, resp.Usage.PromptTokens)
//		ai_convert.SetAIModelOutputToken(ctx, resp.Usage.CompletionTokens)
//		ai_convert.SetAIModelTotalToken(ctx, resp.Usage.TotalTokens)
//	}
//	ai_convert.SetAIStatusNormal(ctx)
//}
