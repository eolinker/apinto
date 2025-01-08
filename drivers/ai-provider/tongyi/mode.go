package tongyi

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
		endPoint: "/compatible-mode/v1/chat/completions",
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

	/*
		错误码对照表
		| HTTP 返回码 | 错误代码 Code            | 错误信息 Message                                                                                         |
		|------------|-------------------------|--------------------------------------------------------------------------------------------------------|
		| 400        | InvalidParameter         | Required parameter(s) missing or invalid, please check the request parameters.                              |
		| 400        | InvalidParameter         | Either "prompt" or "messages" must exist and cannot both be none.                                            |
		| 400        | InvalidParameter         | 'messages' must contain the word 'json' in some form, to use 'response_format' of type 'json_object'.       |
		| 400        | InvalidParameter         | File [id:file-fe-*] format is not supported.                                                             |
		| 400        | DataInspectionFailed     | Input or output data may contain inappropriate content.                                                    |
		| 400        | BadRequest.EmptyInput     | Required input parameter missing from request.                                                            |
		| 400        | BadRequest.EmptyParameters| Required parameter "parameters" missing from request.                                                     |
		| 400        | BadRequest.EmptyModel     | Required parameter "model" missing from request.                                                          |
		| 400        | InvalidURL               | Invalid URL provided in your request.                                                                   |
		| 400        | Arrearage                | Access denied, please make sure your account is in good standing.                                            |
		| 400        | UnsupportedOperation      | The operation is unsupported on the referee object.                                                         |
		| 400        | FlowNotPublished         | Flow has not published yet, please publish flow and try again.                                             |
		| 400        | InvalidSchema            | Database schema is invalid for text2sql.                                                                   |
		| 400        | InvalidSchemaFormat      | Database schema format is invalid for text2sql.                                                            |
		| 400        | FaqRuleBlocked           | Input or output data is blocked by faq rule.                                                              |
		| 400        | CustomRoleBlocked        | Input or output data may contain inappropriate content with custom rule.                                    |
		| 400        | InternalError.Algo       | Missing Content-Length of multimodal url.                                                                  |
		| 400        | invalid_request_error    | Payload Too Large.                                                                                        |
		| 401        | InvalidApiKey            | Invalid API-key provided.                                                                                  |
		| 403        | AccessDenied             | Access denied.                                                                                            |
		| 403        | Workspace.AccessDenied   | Workspace access denied.                                                                                  |
		| 403        | Model.AccessDenied       | Model access denied.                                                                                      |
		| 403        | AccessDenied.Unpurchased | Access to model denied. Please make sure you are eligible for using the model.                              |
		| 404        | WorkSpaceNotFound        | WorkSpace can not be found.                                                                               |
		| 404        | ModelNotFound            | Model can not be found.                                                                                   |
		| 408        | RequestTimeOut           | Request timed out, please try again later.                                                                  |
		| 413        | BadRequest.TooLarge      | Payload Too Large.                                                                                        |
		| 415        | BadRequest.InputDownloadFailed| Failed to download the input file.                                                                      |
		| 415        | BadRequest.UnsupportedFileFormat| Input file format is not supported.                                                                    |
		| 429        | Throttling               | Requests throttling triggered.                                                                              |
		| 429        | Throttling.RateQuota     | Requests rate limit exceeded, please try again later.                                                       |
		| 429        | Throttling.AllocationQuota| Allocated quota exceeded, please increase your quota limit.                                               |
		| 429        | LimitRequests            | You exceeded your current requests list.                                                                    |
		| 429        | Throttling.AllocationQuota| Free allocated quota exceeded.                                                                            |
		| 429        | PrepaidBillOverdue        | The prepaid bill is overdue.                                                                               |
		| 429        | PostpaidBillOverdue       | The postpaid bill is overdue.                                                                               |
		| 429        | CommodityNotPurchased    | Commodity has not purchased yet.                                                                            |
		| 500        | InternalError            | An internal error has occured, please try again later or contact service support.                              |
		| 500        | InternalError.Algo       | An internal error has occured during execution, please try again later or contact service support.           |
		| 500        | SystemError              | An system error has occured, please try again later.                                                         |
		| 500        | InternalError.Timeout    | An internal timeout error has occured during execution, please try again later or contact service support.   |
		| 500        | RewriteFailed            | Failed to rewrite content for prompt.                                                                      |
		| 500        | RetrivalFailed           | Failed to retrieve data from documents.                                                                    |
		| 500        | AppProcessFailed         | Failed to proceed application request.                                                                      |
		| 500        | ModelServiceFailed       | Failed to request model service.                                                                            |
		| 500        | InvokePluginFailed       | Failed to invoke plugin.                                                                                    |
		| 503        | ModelUnavailable         | Model is unavailable, please try again later.                                                                 |
		|            | NetworkError             | Can not find api-key.                                                                                      |
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
	case 401:
		// Handle the invalid API key error.
		convert.SetAIStatusInvalid(ctx)
	case 429:
		switch data.Config.Error.Code {
		case "Throttling", "Throttling.RateQuota", "Throttling.AllocationQuota":
			// Handle the rate limit error.
			convert.SetAIStatusExceeded(ctx)
		default:
			// Handle the insufficient quota error.
			convert.SetAIStatusQuotaExhausted(ctx)
		}
	default:
		convert.SetAIStatusInvalidRequest(ctx)
	}

	responseBody := &convert.ClientResponse{}
	if len(data.Config.Choices) > 0 {
		msg := data.Config.Choices[0]
		responseBody.Message = convert.Message{
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
