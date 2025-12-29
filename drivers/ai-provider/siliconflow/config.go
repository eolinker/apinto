package siliconflow

import (
	ai_convert "github.com/eolinker/apinto/ai-convert"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

func init() {
	driverCreate.Set(ai_convert.ModelTypeChat, func(c *Config) (ai_convert.IConverterDriver, error) {
		return ai_convert.NewOpenAIChat(provider, c.APIKey, c.BaseUrl, 0, nil, errorCallback)
	})
}

func errorCallback(ctx http_service.IHttpContext, body []byte) {
	switch ctx.Response().StatusCode() {

	case 400:
		// Handle the bad request error.
		ai_convert.SetAIStatusInvalidRequest(ctx)
	case 402:
		// Handle the balance is insufficient.
		ai_convert.SetAIStatusQuotaExhausted(ctx)
	case 429:
		// Handle exceed
		ai_convert.SetAIStatusExceeded(ctx)
	case 401:
		// Handle authentication failure
		ai_convert.SetAIStatusInvalid(ctx)
	}
}
