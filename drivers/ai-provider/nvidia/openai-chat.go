package nvidia

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
	// 401	INVALID_API_KEY	The API key is invalid. You can check your API key here: Manage API Key
	// 403	NOT_ENOUGH_BALANCE	Your credit is not enough. You can top up more credit here: Top Up Credit
	// 404	MODEL_NOT_FOUND	The requested model is not found. You can find all the models we support here: https://novita.ai/llm-api or request the Models API to get all available models.
	// 429	RATE_LIMIT_EXCEEDED	You have exceeded the rate limit. Please refer to Rate Limits for more information.
	// 500	MODEL_NOT_AVAILABLE	The requested model is not available now. This is usually due to the model being under maintenance. You can contact us on Discord for more information.
	switch ctx.Response().StatusCode() {
	case 400:
		// Handle the bad request error.
		ai_convert.SetAIStatusInvalidRequest(ctx)
	case 401:
		// Handle the key error.
		ai_convert.SetAIStatusInvalid(ctx)
	case 403:
		// handle credit is exhausted
		ai_convert.SetAIStatusQuotaExhausted(ctx)
	case 429:
		// Handle the rate limit error.
		ai_convert.SetAIStatusExceeded(ctx)
	default:
		ai_convert.SetAIStatusInvalidRequest(ctx)
	}
}
