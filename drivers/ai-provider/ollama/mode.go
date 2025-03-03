package ollama

import (
	"encoding/json"

	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/convert"
	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
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

	// SetProvider the forwarding URI to the chat endpoint.
	httpContext.Proxy().URI().SetPath(c.endPoint)

	// Parse the request body into a base configuration.
	baseCfg := eosc.NewBase[convert.ClientRequest]()
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
	if baseCfg.Config.Stream {
		httpContext.SetLabel("stream", "true")
	}

	// SetProvider the modified body in the HTTP context.
	httpContext.Proxy().Body().SetRaw("application/json", body)
	//httpContext.Response().AppendStreamFunc(c.streamFunc())
	return nil
}

//func (c *Chat) streamFunc() http_context.StreamFunc {
//	return func(ctx http_context.IHttpContext, p []byte) ([]byte, error) {
//		data := eosc.NewBase[Response]()
//		err := json.Unmarshal(p, data)
//		if err != nil {
//			return nil, err
//		}
//		status := ctx.Response().StatusCode()
//		switch status {
//		case 200:
//			// Calculate the token consumption for a successful request.
//			usage := data.Config
//			if usage.Done {
//				convert.SetAIStatusNormal(ctx)
//				convert.SetAIModelInputToken(ctx, usage.PromptEvalCount)
//				convert.SetAIModelOutputToken(ctx, usage.EvalCount)
//				convert.SetAIModelTotalToken(ctx, usage.PromptEvalCount+usage.EvalCount)
//			}
//		case 404:
//			convert.SetAIStatusInvalid(ctx)
//		case 429:
//			convert.SetAIStatusExceeded(ctx)
//		}
//
//		// Prepare the response body for the client.
//		responseBody := &convert.ClientResponse{}
//		resp := data.Config
//		if resp.Message != nil {
//			responseBody.Message = &convert.Message{
//				Role:    resp.Message.Role,
//				Content: resp.Message.Content,
//			}
//			if resp.Done {
//				responseBody.FinishReason = convert.FinishStop
//			}
//		} else {
//			responseBody.Code = -1
//			responseBody.Error = "response message is nil"
//		}
//
//		// Marshal the modified response body back into JSON.
//		body, err := json.Marshal(responseBody)
//		if err != nil {
//			return nil, err
//		}
//		return body, nil
//	}
//}

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
	if body == nil {
		return nil
	}

	// Parse the response body into a base configuration.
	data := eosc.NewBase[convert.Response]()
	err = json.Unmarshal(body, data)
	if err != nil {
		return err
	}
	switch httpContext.Response().StatusCode() {
	case 200:
		// Calculate the token consumption for a successful request.
		usage := data.Config.Usage
		convert.SetAIStatusNormal(ctx)
		convert.SetAIModelInputToken(ctx, usage.PromptTokens)
		convert.SetAIModelOutputToken(ctx, usage.CompletionTokens)
		convert.SetAIModelTotalToken(ctx, usage.TotalTokens)
	}
	//
	//// Prepare the response body for the client.
	//responseBody := &convert.ClientResponse{}
	//resp := data.Config
	//if resp.Choices != nil {
	//	responseBody.Message = &convert.Message{
	//		Role:    resp.Message.Role,
	//		Content: resp.Message.Content,
	//	}
	//	responseBody.FinishReason = convert.FinishStop
	//} else {
	//	responseBody.Code = -1
	//	responseBody.Error = resp.Error
	//}
	//
	//// Marshal the modified response body back into JSON.
	//body, err = json.Marshal(responseBody)
	//if err != nil {
	//	return err
	//}
	//
	//httpContext.Response().SetBody(body)

	// SetProvider the modified response in the HTTP context.
	return nil
}
