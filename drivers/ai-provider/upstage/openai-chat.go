package upstage

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
	switch ctx.Response().StatusCode() {
	case 400:
		// Handle the bad request error.
		ai_convert.SetAIStatusInvalidRequest(ctx)
	case 403:
		// Handle the quota exhausted error.
		ai_convert.SetAIStatusQuotaExhausted(ctx)
	case 401:
		// Handle the invalid key error.
		ai_convert.SetAIStatusInvalid(ctx)
	case 429:
		// Handle the rate limit exceeded error.
		ai_convert.SetAIStatusExceeded(ctx)
	default:
		// Handle the unknown error.
		ai_convert.SetAIStatusInvalidRequest(ctx)

	}
}
