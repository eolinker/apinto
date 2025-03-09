package spark

import (
	"encoding/json"

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
		endPoint: "/v1/chat/completions",
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
	messages := make([]Message, 0, len(baseCfg.Config.Messages)+1)
	for _, m := range baseCfg.Config.Messages {
		messages = append(messages, Message{
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
		if data.Config.Code == 0 {
			// Calculate the token consumption for a successful request.
			usage := data.Config.Usage
			ai_convert.SetAIStatusNormal(ctx)
			ai_convert.SetAIModelInputToken(ctx, usage.PromptTokens)
			ai_convert.SetAIModelOutputToken(ctx, usage.CompletionTokens)
			ai_convert.SetAIModelTotalToken(ctx, usage.TotalTokens)
		}
	case 400:
		// Handle the bad request error.
		ai_convert.SetAIStatusInvalidRequest(ctx)
	case 401:
		// Handle authentication failure
		ai_convert.SetAIStatusInvalid(ctx)
	}
	if data.Config.Error != nil {
		handleErrorCode(ctx, data.Config.Error.Code)
	}
	responseBody := &ai_convert.ClientResponse{}
	if len(data.Config.Choices) > 0 {
		msg := data.Config.Choices[0]
		responseBody.Message = &ai_convert.Message{
			Role:    msg.Message.Role,
			Content: msg.Message.Content,
		}
		responseBody.FinishReason = "stop"
	} else {
		responseBody.Code = -1
		if data.Config.Error != nil {
			responseBody.Error = data.Config.Error.Message
		} else {
			responseBody.Error = "no response"
		}
	}
	body, err = json.Marshal(responseBody)
	if err != nil {
		return err
	}
	httpContext.Response().SetBody(body)
	return nil
}

/**
 * @description: Handle the error code returned by the AI provider.
 */
func handleErrorCode(ctx eocontext.EoContext, errorCode interface{}) {
	switch errorCode {
	case "11200":
		// Handle the insufficient quota error.
		ai_convert.SetAIStatusQuotaExhausted(ctx)
	case "10007", "11201", "11202", "11203":
		// Handle the rate limit error.
		ai_convert.SetAIStatusExceeded(ctx)
	default:
		ai_convert.SetAIStatusInvalidRequest(ctx)
	}
}
