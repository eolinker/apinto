package hunyuan

import (
	"encoding/json"
	"fmt"

	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/convert"
	ai_provider "github.com/eolinker/apinto/drivers/ai-provider"
	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
)

type FNewModelMode func(string) IModelMode

var (
	modelModes = map[string]FNewModelMode{
		ai_provider.ModeChat.String(): NewChat,
	}
)

type IModelMode interface {
	Endpoint() string
	convert.IConverter
}

type Chat struct {
	endPoint string
}

func NewChat(model string) IModelMode {
	return &Chat{
		endPoint: "/",
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
	msgLen := len(baseCfg.Config.Messages)
	if msgLen < 1 {
		return fmt.Errorf("message length is less than 1")
	}
	if msgLen == 1 {
		messages = append(messages, Message{
			Role:    "user",
			Content: baseCfg.Config.Messages[0].Content,
		})
	} else {
		messages = append(messages, Message{
			Role:    "user",
			Content: baseCfg.Config.Messages[0].Content,
		})
		for _, m := range baseCfg.Config.Messages {
			messages = append(messages, Message{
				Role:    m.Role,
				Content: m.Content,
			})
		}
	}

	baseCfg.SetAppend("Messages", messages)
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
	if httpContext.Response().StatusCode() != 200 {
		return nil
	}
	body := httpContext.Response().GetBody()
	data := eosc.NewBase[Response]()
	err = json.Unmarshal(body, data)
	if err != nil {
		return err
	}
	responseBody := &ai_provider.ClientResponse{}
	if len(data.Config.Response.Choices) > 0 {
		//if data.Config.Object == "chat.completion" {
		msg := data.Config.Response.Choices[0]
		responseBody.Message = ai_provider.Message{
			Role:    msg.Message.Role,
			Content: msg.Message.Content,
		}
		responseBody.FinishReason = msg.FinishReason
		//}
	} else {
		responseBody.Code = -1
		responseBody.Error = data.Config.Response.Error.Message
	}

	body, err = json.Marshal(responseBody)
	if err != nil {
		return err
	}
	httpContext.Response().SetBody(body)
	return nil
}
