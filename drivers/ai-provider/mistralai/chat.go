package mistralai

import (
	ai_convert "github.com/eolinker/apinto/ai-convert"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

func init() {
	driverCreate.Set(ai_convert.ModelTypeChat, func(c *Config) (ai_convert.IConverterDriver, error) {
		return ai_convert.NewOpenAIChat(provider, "", c.BaseUrl, 0, nil, errorCallback)
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
	case 422:
		// Handle the bad request error.
		ai_convert.SetAIStatusInvalidRequest(ctx)
	case 401:
		// Handle the invalid API key error.
		ai_convert.SetAIStatusInvalid(ctx)
	case 403:
		ai_convert.SetAIStatusQuotaExhausted(ctx)
	case 429:
		ai_convert.SetAIStatusExceeded(ctx)
	default:
		ai_convert.SetAIStatusInvalidRequest(ctx)
	}
}
