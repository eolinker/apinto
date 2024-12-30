package fireworks

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
		endPoint: "/inference/v1/chat/completions",
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

	// 400	Bad Request	Invalid input or malformed request.	Review the request parameters and ensure they match the expected format.
	// 401	Unauthorized	Invalid API key or insufficient permissions.	Verify your API key and ensure it has the correct permissions.
	// 402	Payment Required	User’s account is not on a paid plan or has exceeded usage limits.	Check your billing status and ensure your payment method is up to date. Upgrade your plan if necessary.
	// 403	Forbidden	The model name may be incorrect, or the model does not exist. This error is also returned to avoid leaking information about model availability.	Verify the model name on the Fireworks site and ensure it exists. Double-check the spelling of the model name in your request.
	// 404	Not Found	The API endpoint is incorrect, or the resource path is invalid (e.g., a user tried accessing /v1/foobar instead of a valid endpoint).	Verify the URL path in your request and ensure you are using the correct API endpoint as per the documentation.
	// 405	Method Not Allowed	Using an unsupported HTTP method (e.g., using GET instead of POST).	Check the API documentation for the correct HTTP method to use for the request.
	// 408	Request Timeout	The request took too long to complete, possibly due to server overload or network issues.	Retry the request after a brief wait. Consider increasing the timeout value if applicable.
	// 412	Precondition Failed	This error occurs when attempting to invoke a LoRA model that failed to load. The final validation of the model happens during inference, not at upload time.	Check the body of the request for a detailed error message. Ensure the LoRA model was uploaded correctly and is compatible. Contact support if the issue persists.
	// 413	Payload Too Large	Input data exceeds the allowed size limit.	Reduce the size of the input payload (e.g., by trimming large text or image data).
	// 429	Over Quota	The user has reached the API rate limit.	Wait for the quota to reset or upgrade your plan for a higher rate limit.
	// 500	Internal Server Error	This indicates a server-side code bug and is unlikely to resolve on its own.	Contact Fireworks support immediately, as this error typically requires intervention from the engineering team.
	// 502	Bad Gateway	The server received an invalid response from an upstream server.	Wait and retry the request. If the error persists, it may indicate a server outage.
	// 503	Service Unavailable	The service is down for maintenance or experiencing issues.	Retry the request after some time. Check for any maintenance announcements.
	// 504	Gateway Timeout	The server did not receive a response in time from an upstream server.	Wait briefly and retry the request. Consider using a shorter input prompt if applicable.
	// 520	Unknown Error	An unexpected error occurred with no clear explanation.	Retry the request. If the issue persists, contact support for further assistance.
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
	case 401, 403:
		// Handle the Invalid API key error. 官方返回状态码与文档不一致，应该返回401，实际返回403
		convert.SetAIStatusInvalid(ctx)
	case 402:
		// Handle the insufficient quota error.
		convert.SetAIStatusQuotaExhausted(ctx)
	case 429:
		// Handle the rate limit error.
		convert.SetAIStatusExceeded(ctx)
	default:
		convert.SetAIStatusInvalidRequest(ctx)
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
