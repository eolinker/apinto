package openrouter

import (
	"encoding/json"

	"github.com/eolinker/apinto/convert"
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
		endPoint: "/api/v1/chat/completions",
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
	baseCfg := eosc.NewBase[convert.ClientRequest]()
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

	// 400: Bad Request (invalid or missing params, CORS)
	// 401: Invalid credentials (OAuth session expired, disabled/invalid API key)
	// 402: Your account or API key has insufficient credits. Add more credits and retry the request.
	// 403: Your chosen model requires moderation and your input was flagged
	// 408: Your request timed out
	// 429: You are being rate limited
	// 502: Your chosen model is down or we received an invalid response from it
	// 503: There is no available model provider that meets your routing requirements
	switch httpContext.Response().StatusCode() {
	case 200:
		if data.Config.Error != nil {
			// Handle the error response.
			switch data.Config.Error.Code {
			case 400:
				convert.SetAIStatusInvalidRequest(ctx)
			case 401:
				convert.SetAIStatusInvalid(ctx)
			case 402:
				convert.SetAIStatusQuotaExhausted(ctx)
			case 429:
				convert.SetAIStatusExceeded(ctx)
			default:
				convert.SetAIStatusInvalidRequest(ctx)
			}
		} else {
			// Calculate the token consumption for a successful request.
			usage := data.Config.Usage
			convert.SetAIStatusNormal(ctx)
			convert.SetAIModelInputToken(ctx, usage.PromptTokens)
			convert.SetAIModelOutputToken(ctx, usage.CompletionTokens)
			convert.SetAIModelTotalToken(ctx, usage.TotalTokens)
		}
	case 400:
		// Handle the bad request error.
		convert.SetAIStatusInvalidRequest(ctx)
	case 401:
		// Handle the invalid key error.
		convert.SetAIStatusInvalid(ctx)
	case 402:
		// Handle the expired key error.
		convert.SetAIStatusQuotaExhausted(ctx)
	case 429:
		convert.SetAIStatusExceeded(ctx)
	default:
		convert.SetAIStatusInvalidRequest(ctx)
	}

	responseBody := &convert.ClientResponse{}
	if len(data.Config.Choices) > 0 {
		msg := data.Config.Choices[0]
		responseBody.Message = &convert.Message{
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
