package perfxcloud

import (
	ai_convert "github.com/eolinker/apinto/ai-convert"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

func init() {
	driverCreate.Set(ai_convert.ModelTypeOpenAIChat, func(mt ai_convert.ModelType, c *Config) (ai_convert.IConverterDriver, error) {
		return ai_convert.NewOpenAIChat(provider, c.APIKey, c.BaseUrl, mt, 0, nil, errorCallback)
	})
}

func errorCallback(ctx http_service.IHttpContext, body []byte) {
	// 400: Bad Request (invalid or missing params, CORS)
	// 401: Invalid credentials (OAuth session expired, disabled/invalid API key)
	// 402: Your account or API key has insufficient credits. Add more credits and retry the request.
	// 403: Your chosen model requires moderation and your input was flagged
	// 408: Your request timed out
	// 429: You are being rate limited
	// 502: Your chosen model is down or we received an invalid response from it
	// 503: There is no available model provider that meets your routing requirements
	switch ctx.Response().StatusCode() {
	case 400:
		// Handle the bad request error.
		ai_convert.SetAIStatusInvalidRequest(ctx)
	case 401:
		// Handle the invalid key error.
		ai_convert.SetAIStatusInvalid(ctx)
	case 402:
		// Handle the expired key error.
		ai_convert.SetAIStatusQuotaExhausted(ctx)
	case 429:
		ai_convert.SetAIStatusExceeded(ctx)
	default:
		ai_convert.SetAIStatusInvalidRequest(ctx)
	}
}
