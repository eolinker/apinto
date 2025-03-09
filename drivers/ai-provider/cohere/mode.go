package cohere

import (
	"encoding/json"
	"fmt"

	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
)

var (
	modelModes = map[string]IModelMode{
		ai_convert.ModeChat.String(): NewChat(),
	}
)

type IModelMode interface {
	Endpoint() string
	ai_convert.IConverter
}

type Chat struct {
	endPoint string
}

func NewChat() *Chat {
	return &Chat{
		endPoint: "/v2/chat",
	}
}

func (c *Chat) Endpoint() string {
	return c.endPoint
}

func (c *Chat) RequestConvert(ctx eocontext.EoContext, extender map[string]interface{}) error {
	httpContext, err := http_context.Assert(ctx)
	if err != nil {
		return err
	}
	body, err := httpContext.Proxy().Body().RawBody()
	if err != nil {
		return err
	}
	// 设置转发地址
	httpContext.Proxy().URI().SetPath(c.endPoint)
	baseCfg := eosc.NewBase[ai_convert.ClientRequest]()
	err = json.Unmarshal(body, baseCfg)
	if err != nil {
		return err
	}
	messages := make([]RequestMessage, 0, len(baseCfg.Config.Messages)+1)
	for _, m := range baseCfg.Config.Messages {
		messages = append(messages, RequestMessage{
			Role:    m.Role,
			Content: m.Content,
		})
	}
	baseCfg.SetAppend("messages", messages)
	for k, v := range extender {
		baseCfg.SetAppend(k, v)
	}
	body, err = json.Marshal(baseCfg)
	if err != nil {
		return err
	}
	httpContext.Proxy().Body().SetRaw("application/json", body)
	return nil
}

func (c *Chat) ResponseConvert(ctx eocontext.EoContext) error {
	httpContext, err := http_context.Assert(ctx)
	if err != nil {
		return err
	}
	body := httpContext.Response().GetBody()
	data := eosc.NewBase[Response]()
	err = json.Unmarshal(body, data)
	if err != nil {
		return err
	}
	// 针对不同响应做出处理
	switch httpContext.Response().StatusCode() {
	case 200:
		// Calculate the token consumption for a successful request.
		usage := data.Config.Usage
		ai_convert.SetAIStatusNormal(ctx)
		ai_convert.SetAIModelInputToken(ctx, usage.Tokens.InputTokens)
		ai_convert.SetAIModelOutputToken(ctx, usage.Tokens.OutputTokens)
		// 待定
		ai_convert.SetAIModelTotalToken(ctx, usage.BilledUnits.InputTokens+usage.Tokens.OutputTokens)
	case 400, 422:
		// Handle the bad request error.
		ai_convert.SetAIStatusInvalidRequest(ctx)
	case 402:
		// Handle the balance is insufficient.
		ai_convert.SetAIStatusQuotaExhausted(ctx)
	case 429:
		// Handle exceed
		ai_convert.SetAIStatusExceeded(ctx)
	case 401:
		// Handle authentication failure
		ai_convert.SetAIStatusInvalid(ctx)
	}
	responseBody := &ai_convert.ClientResponse{}
	if data.Config.Id != "" {
		switch tmp := data.Config.Message.(type) {
		case map[string]interface{}:
			{
				responseMessage := ai_convert.MapToStruct[ResponseMessage](tmp)
				responseBody.Message = &ai_convert.Message{
					Role:    responseMessage.Role,
					Content: responseMessage.Content[0].Text,
				}
				responseBody.FinishReason = data.Config.FinishReason
			}
		default:
			return fmt.Errorf("failed to convert response message")
		}
	} else {
		responseBody.Code = -1
		responseBody.Error = data.Config.Message.(string)
	}
	body, err = json.Marshal(responseBody)
	if err != nil {
		return err
	}
	httpContext.Response().SetBody(body)
	return nil
}
