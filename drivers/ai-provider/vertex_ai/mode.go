package vertex_ai

import (
	"encoding/json"

	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/convert"
	ai_provider "github.com/eolinker/apinto/drivers/ai-provider"
	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
)

type FNewModelMode func(string) convert.IChildConverter

var (
	modelModes = map[string]FNewModelMode{
		ai_provider.ModeChat.String(): NewChat,
	}
)

type Chat struct {
	model    string
	endPoint string
}

func NewChat(model string) convert.IChildConverter {
	return &Chat{
		endPoint: "/v1/projects/%s/locations/%s/publishers/google/models/%s:generateContent",
		model:    model,
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
	baseCfg := eosc.NewBase[ai_provider.ClientRequest]()
	err = json.Unmarshal(body, baseCfg)
	if err != nil {
		return err
	}
	messages := make([]Content, 0, len(baseCfg.Config.Messages)+1)
	for _, m := range baseCfg.Config.Messages {
		role := "user"
		if m.Role == "system" && len(baseCfg.Config.Messages) > 1 {
			role = "model"
		}
		parts := make([]map[string]interface{}, 0, 1)
		if m.Content != "" {
			parts = append(parts, map[string]interface{}{
				"text": m.Content,
			})
		}
		messages = append(messages, Content{
			Role:  role,
			Parts: parts,
		})
	}
	baseCfg.SetAppend("contents", messages)
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
	if len(data.Config.Candidates) > 0 {
		msg := data.Config.Candidates[0]
		role := "user"
		if msg.Content.Role == "model" {
			role = "assistant"
		}
		text := ""
		if len(msg.Content.Parts) > 0 {
			if v, ok := msg.Content.Parts[0]["text"]; ok {
				text = v.(string)
			}
		}

		responseBody.Message = ai_provider.Message{
			Role:    role,
			Content: text,
		}
		responseBody.FinishReason = msg.FinishReason
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
