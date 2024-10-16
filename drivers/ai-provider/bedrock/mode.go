package bedrock

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
		endPoint: fmt.Sprintf("/model/%s/converse", model),
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
	messages := make([]Message, 0, len(baseCfg.Config.Messages))
	systemMessage := make([]Content, 0)

	for _, m := range baseCfg.Config.Messages {
		if m.Role == "system" {
			systemMessage = append(systemMessage, Content{Text: m.Content})
		} else {
			messages = append(messages, Message{
				Role:    m.Role,
				Content: []*Content{{Text: m.Content}},
			})
		}
	}
	baseCfg.SetAppend("messages", messages)
	baseCfg.SetAppend("system", systemMessage)

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
	if data.Config.Output.Message != nil && len(data.Config.Output.Message.Content) > 0 {
		msg := data.Config.Output.Message
		responseBody.Message = ai_provider.Message{
			Role:    msg.Role,
			Content: msg.Content[0].Text,
		}
		responseBody.FinishReason = data.Config.StopReason
	} else {
		responseBody.Code = -1
		responseBody.Error = "no response"
	}
	body, err = json.Marshal(responseBody)
	if err != nil {
		return err
	}
	httpContext.Response().SetBody(body)
	return nil
}
