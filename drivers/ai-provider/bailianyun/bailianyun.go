package bailianyun

import (
	"encoding/json"
	"fmt"
	"net/url"

	ai_convert "github.com/eolinker/apinto/ai-convert"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
)

func init() {
	ai_convert.RegisterConverterCreateFunc("bailian", Create)
}

type Config struct {
	APIKey  string `json:"api_key"`
	BaseUrl string `json:"base_url"`
}

// checkConfig validates the provided configuration.
// It ensures the required fields are set and checks the validity of the Base URL if provided.
//
// Parameters:
//   - v: An interface{} expected to be a pointer to a Config struct.
//
// Returns:
//   - *Config: The validated configuration cast to *Config.
//   - error: An error if the validation fails, or nil if it succeeds.
func checkConfig(conf *Config) error {
	// Check if the APIKey is provided. It is a required field.
	if conf.APIKey == "" {
		return fmt.Errorf("api_key is required")
	}
	if conf.BaseUrl != "" {
		u, err := url.Parse(conf.BaseUrl)
		if err != nil {
			// Return an error if the Base URL cannot be parsed.
			return fmt.Errorf("base url is invalid")
		}
		// Ensure the parsed URL contains both a scheme and a host.
		if u.Scheme == "" || u.Host == "" {
			return fmt.Errorf("base url is invalid")
		}
	}
	return nil
}

func Create(cfg string) (ai_convert.IConverter, error) {
	var conf Config
	err := json.Unmarshal([]byte(cfg), &conf)
	if err != nil {
		return nil, err
	}
	err = checkConfig(&conf)
	if err != nil {
		return nil, err
	}

	return ai_convert.NewOpenAIConvert(conf.APIKey, conf.BaseUrl, 0, errorCallback)
}

func errorCallback(ctx http_service.IHttpContext, body []byte) {
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

	switch ctx.Response().StatusCode() {
	case 400:
		// Handle the bad request error.
		ai_convert.SetAIStatusInvalidRequest(ctx)
	case 401:
		// Handle the invalid API key error.
		ai_convert.SetAIStatusInvalid(ctx)
	case 429:
		var data ai_convert.Response
		err := json.Unmarshal(body, &data)
		if err != nil {
			log.Errorf("unmarshal body error: %v, body: %s", err, string(body))
			return
		}
		switch data.Error.Code {
		case "Throttling", "Throttling.RateQuota", "Throttling.AllocationQuota":
			// Handle the rate limit error.
			ai_convert.SetAIStatusExceeded(ctx)
		default:
			// Handle the insufficient quota error.
			ai_convert.SetAIStatusQuotaExhausted(ctx)
		}
	default:
		ai_convert.SetAIStatusInvalidRequest(ctx)
	}
}
