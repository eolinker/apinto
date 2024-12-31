package nvidia

import (
	"encoding/json"
	"fmt"

	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/convert"
	ai_provider "github.com/eolinker/apinto/drivers/ai-provider"
	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
)

var (
	modelModes = map[string]IModelMode{
		ai_provider.ModeChat.String(): NewChat(),
	}
)

type IModelMode interface {
	Endpoint() string
	convert.IConverter
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
	baseCfg := eosc.NewBase[ai_provider.ClientRequest]()
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
	var data *eosc.Base[Response]
	var errData *eosc.Base[ErrorResponse] = nil
	if httpContext.Response().StatusCode() == 200 {
		data = eosc.NewBase[Response]()
		err = json.Unmarshal(body, data)
	} else {
		errData = eosc.NewBase[ErrorResponse]()
		err = json.Unmarshal(body, errData)
	}
	if err != nil {
		return err
	}
	// 针对不同响应做出处理
	switch httpContext.Response().StatusCode() {
	case 200:
		// Calculate the token consumption for a successful request.
		usage := data.Config.Usage
		ai_provider.SetAIStatusNormal(ctx)
		ai_provider.SetAIModelInputToken(ctx, usage.PromptTokens)
		ai_provider.SetAIModelOutputToken(ctx, usage.CompletionTokens)
		ai_provider.SetAIModelTotalToken(ctx, usage.TotalTokens)
	case 400, 422, 403:
		// Handle the bad request error.
		ai_provider.SetAIStatusInvalidRequest(ctx)
	case 429:
		// Handle exceed
		ai_provider.SetAIStatusExceeded(ctx)
	case 401:
		// Handle authentication failure
		ai_provider.SetAIStatusInvalid(ctx)
	}
	responseBody := &ai_provider.ClientResponse{}
	if data != nil && len(data.Config.Choices) > 0 {
		msg := data.Config.Choices[0]
		responseBody.Message = ai_provider.Message{
			Role:    msg.Message.Role,
			Content: msg.Message.Content,
		}
		responseBody.FinishReason = msg.FinishReason
	} else {
		responseBody.Code = -1
		if errData != nil {
			responseBody.Error = fmt.Sprintf("%s: %s", errData.Config.Title, errData.Config.Detail)
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
