package moonshot

import (
	"encoding/json"
	ai_convert "github.com/eolinker/apinto/ai-convert"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
)

func init() {
	driverCreate.Set(ai_convert.ModelTypeChat, func(c *Config) (ai_convert.IConverterDriver, error) {
		return ai_convert.NewOpenAIChat(provider, c.APIKey, c.BaseUrl, 0, nil, errorCallback)
	})

}

func errorCallback(ctx http_service.IHttpContext, body []byte) {
	// HTTP Status Codes for Moonshot API
	// Status Code | Type                | Error Message
	// ------------|---------------------|-------------------------------------
	// 200         | Success             | Request was successful.
	// 400         | Client Error        | Invalid request parameters (invalid_request_error).
	// 401         | Authentication Error | Invalid API key (invalid_key).
	// 403         | Forbidden           | Access denied (forbidden_error).
	// 404         | Not Found           | Resource not found (not_found_error).
	// 429         | Rate Limit Exceeded | Too many requests (rate_limit_error).
	// 500         | Server Error        | Internal server error (server_error).
	// 503         | Service Unavailable  | Service is temporarily unavailable (service_unavailable).
	switch ctx.Response().StatusCode() {

	case 400:
		// Handle the bad request error.
		ai_convert.SetAIStatusInvalidRequest(ctx)
	case 401:
		// 过期和无效的API密钥
		ai_convert.SetAIStatusInvalid(ctx)
	case 429:
		var data ai_convert.Response
		err := json.Unmarshal(body, &data)
		if err != nil {
			log.Errorf("unmarshal body error: %v, body: %s", err, body)
			return
		}
		switch data.Error.Type {
		case "exceeded_current_quota_error":
			// Handle the insufficient quota error.
			ai_convert.SetAIStatusQuotaExhausted(ctx)
		case "engine_overloaded_error", "rate_limit_reached_error":
			// Handle the rate limit error.
			ai_convert.SetAIStatusExceeded(ctx)
		}
	default:
		ai_convert.SetAIStatusInvalidRequest(ctx)
	}
}
