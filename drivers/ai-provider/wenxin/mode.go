package wenxin

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/eolinker/eosc"

	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
)

type FNewModelMode func(string) IModelMode

var (
	modelModes = map[string]FNewModelMode{
		ai_convert.ModeChat.String(): NewChat,
	}
)

type IModelMode interface {
	Endpoint() string
	ai_convert.IConverter
}

type Chat struct {
	endPoint string
}

func NewChat(model string) IModelMode {
	endPoint := fmt.Sprintf("/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/%s", model)
	if model == "ernie-4.0-8k" {
		endPoint = "/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/completions_pro"
	}

	return &Chat{
		endPoint: endPoint,
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
	var tmpMsg []*ai_convert.Message
	msgLen := len(baseCfg.Config.Messages)
	if msgLen != 0 && msgLen%2 == 0 {
		// 合并第一第二条信息
		firstMsg := strings.Builder{}
		firstMsg.WriteString(baseCfg.Config.Messages[0].Content + "\n")
		firstMsg.WriteString(baseCfg.Config.Messages[1].Content)
		messages = append(messages, Message{
			Role:    "user",
			Content: firstMsg.String(),
		})
		tmpMsg = baseCfg.Config.Messages[2:]
	} else {
		messages = append(messages, Message{
			Role:    "user",
			Content: baseCfg.Config.Messages[0].Content,
		})
		tmpMsg = baseCfg.Config.Messages[1:]
	}
	for _, m := range tmpMsg {
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

	switch httpContext.Response().StatusCode() {
	case 200:
		// Calculate the token consumption for a successful request.
		if data.Config.ErrorCode != 0 {
			switch data.Config.ErrorCode {
			case 17, 19:
				// Handle the insufficient quota error.
				ai_convert.SetAIStatusQuotaExhausted(ctx)
			case 4, 18, 336501, 336502, 336503, 336504, 336505, 336507:
				// Handle the rate limit error.
				ai_convert.SetAIStatusExceeded(ctx)
			case 13, 14, 100, 110, 111:
				// Handle the invalid token error.
				ai_convert.SetAIStatusInvalid(ctx)
			default:
				ai_convert.SetAIStatusInvalidRequest(ctx)
			}
		} else {
			usage := data.Config.Usage
			ai_convert.SetAIStatusNormal(ctx)
			ai_convert.SetAIModelInputToken(ctx, usage.PromptTokens)
			ai_convert.SetAIModelOutputToken(ctx, usage.CompletionTokens)
			ai_convert.SetAIModelTotalToken(ctx, usage.TotalTokens)
		}
	case 400:
		// Handle the bad request error.
		ai_convert.SetAIStatusInvalidRequest(ctx)
	case 403:
		// Handle the invalid token error.
		ai_convert.SetAIStatusInvalid(ctx)
	default:
		ai_convert.SetAIStatusInvalidRequest(ctx)
	}

	responseBody := &ai_convert.ClientResponse{}
	if data.Config.ErrorCode == 0 {
		//if data.Config.Object == "chat.completion" {
		msg := data.Config
		responseBody.Message = &ai_convert.Message{
			Role:    "assistant",
			Content: msg.Result,
		}
		responseBody.FinishReason = msg.FinishReason
		//}
	} else {
		responseBody.Code = data.Config.ErrorCode
		responseBody.Error = data.Config.ErrorMsg
	}

	body, err = json.Marshal(responseBody)
	if err != nil {
		return err
	}
	httpContext.Response().SetBody(body)
	return nil
}
