package yi

import (
	"encoding/json"

	"github.com/eolinker/apinto/convert"
	ai_provider "github.com/eolinker/apinto/drivers/ai-provider"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
)

var (
	modelModes = map[string]IModelMode{
		convert.ModeChat.String(): NewChat(),
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
	data := eosc.NewBase[Response]()
	err = json.Unmarshal(body, data)
	if err != nil {
		return err
	}
	/*
		错误码对照表
		| HTTP 返回码 | 错误代码     | 原因                                   | 解决方案                             |
		|------------|------------|----------------------------------------|--------------------------------------|
		| 400        | Bad request | 模型的输入+输出（max_tokens）超过了模型的最大上下文。 | 减少模型的输入，或将 max_tokens 参数值设置更小。 |
		|            |            | 输入格式错误。                           | 检查输入格式，确保正确。例如，模型名必须全小写，yi-lightning。 |
		| 401        | Authentication Error | API Key缺失或无效。             | 请确保你的 API Key 有效。               |
		| 404        | Not found   | 无效的 Endpoint URL 或模型名。           | 确保使用正确的 Endpoint URL 或模型名。   |
		| 429        | Too Many Requests | 在短时间内发出的请求太多。         | 控制请求速率。                       |
		| 500        | Internal Server Error | 服务端内部错误。           | 请稍后重试。                         |
		| 529        | System busy | 系统繁忙，请重试。                     | 请 1 分钟后重试。                     |
	*/
	switch httpContext.Response().StatusCode() {
	case 200:
		// Calculate the token consumption for a successful request.
		usage := data.Config.Usage
		convert.SetAIStatusNormal(ctx)
		convert.SetAIModelInputToken(ctx, usage.PromptTokens)
		convert.SetAIModelOutputToken(ctx, usage.CompletionTokens)
		convert.SetAIModelTotalToken(ctx, usage.TotalTokens)
	case 400:
		// Handle the bad request error.
		convert.SetAIStatusInvalidRequest(ctx)
	case 429:
		convert.SetAIStatusExceeded(ctx)
	case 401:
		// Handle the authentication error.
		convert.SetAIStatusInvalid(ctx)
	}
	responseBody := &ai_provider.ClientResponse{}
	if len(data.Config.Choices) > 0 {
		msg := data.Config.Choices[0]
		responseBody.Message = ai_provider.Message{
			Role:    msg.Message.Role,
			Content: msg.Message.Content,
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
