package openAI

import (
	"encoding/json"

	"github.com/eolinker/apinto/encoder"

	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/convert"
	ai_provider "github.com/eolinker/apinto/drivers/ai-provider"
	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
)

// modelModes defines the available model modes and their corresponding implementations.
var (
	modelModes = map[string]IModelMode{
		ai_provider.ModeChat.String(): NewChat(),
	}
)

// IModelMode defines the interface for model modes.
// It includes methods for retrieving the endpoint and performing data conversion.
type IModelMode interface {
	Endpoint() string
	convert.IConverter
}

// Chat represents the chat mode implementation.
// It includes the endpoint and methods for request and response conversion.
type Chat struct {
	endPoint string // The API endpoint for chat completions.
}

// NewChat initializes and returns a new Chat instance.
func NewChat() *Chat {
	return &Chat{
		endPoint: "/v1/chat/completions",
	}
}

// Endpoint returns the API endpoint for the Chat mode.
func (c *Chat) Endpoint() string {
	return c.endPoint
}

// RequestConvert converts the request body for the Chat mode.
// It modifies the HTTP context to include the appropriate endpoint and formatted request body.
func (c *Chat) RequestConvert(ctx eocontext.EoContext, extender map[string]interface{}) error {
	// Assert the context as an HTTP context.
	httpContext, err := http_context.Assert(ctx)
	if err != nil {
		return err
	}

	// Retrieve the raw request body.
	body, err := httpContext.Proxy().Body().RawBody()
	if err != nil {
		return err
	}

	// Set the forwarding URI to the chat endpoint.
	httpContext.Proxy().URI().SetPath(c.endPoint)

	// Parse the request body into a base configuration.
	baseCfg := eosc.NewBase[ai_provider.ClientRequest]()
	err = json.Unmarshal(body, baseCfg)
	if err != nil {
		return err
	}

	// Convert messages and append them to the configuration.
	messages := make([]Message, 0, len(baseCfg.Config.Messages)+1)
	for _, m := range baseCfg.Config.Messages {
		messages = append(messages, Message{
			Role:    m.Role,
			Content: m.Content,
		})
	}
	baseCfg.SetAppend("messages", messages)

	// Append additional fields from the extender.
	for k, v := range extender {
		baseCfg.SetAppend(k, v)
	}

	// Marshal the updated configuration back into JSON.
	body, err = json.Marshal(baseCfg)
	if err != nil {
		return err
	}

	// Set the modified body in the HTTP context.
	httpContext.Proxy().Body().SetRaw("application/json", body)

	return nil
}

// ResponseConvert converts the response body for the Chat mode.
// It processes the response to ensure it conforms to the expected format and encoding.
func (c *Chat) ResponseConvert(ctx eocontext.EoContext) error {
	// Assert the context as an HTTP context.
	httpContext, err := http_context.Assert(ctx)
	if err != nil {
		return err
	}

	// Retrieve the response body.
	body := httpContext.Response().GetBody()

	// Check the content encoding and convert to UTF-8 if necessary.
	encoding := httpContext.Response().Headers().Get("content-encoding")
	if encoding != "utf-8" && encoding != "" {
		body, err = encoder.ToUTF8(encoding, body)
		if err != nil {
			return err
		}
	}

	// Parse the response body into a base configuration.
	data := eosc.NewBase[Response]()
	err = json.Unmarshal(body, data)
	if err != nil {
		return err
	}
	switch httpContext.Response().StatusCode() {
	case 200:
		// Calculate the token consumption for a successful request.
		usage := data.Config.Usage
		ai_provider.SetAIStatusNormal(ctx)
		ai_provider.SetAIModelInputToken(ctx, usage.PromptTokens)
		ai_provider.SetAIModelOutputToken(ctx, usage.CompletionTokens)
		ai_provider.SetAIModelTotalToken(ctx, usage.TotalTokens)
	case 400:
		// Handle the bad request error.
		ai_provider.SetAIStatusInvalidRequest(ctx)
	case 429:
		switch data.Config.Error.Type {
		case "insufficient_quota":
			// Handle the insufficient quota error.
			ai_provider.SetAIStatusQuotaExhausted(ctx)
		case "rate_limit_error":
			// Handle the rate limit error.
			ai_provider.SetAIStatusExceeded(ctx)
		}
	case 401:
		// 过期和无效的API密钥
		ai_provider.SetAIStatusInvalid(ctx)
	}

	// Prepare the response body for the client.
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
		responseBody.Error = data.Config.Error.Message
	}

	// Marshal the modified response body back into JSON.
	body, err = json.Marshal(responseBody)
	if err != nil {
		return err
	}

	// Set the updated body and encoding in the HTTP context.
	httpContext.Response().SetHeader("content-encoding", "utf-8")
	httpContext.Response().SetBody(body)

	return nil
}
