package wenxin

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
	//case 200:
	//	// Calculate the token consumption for a successful request.
	//	if data.Config.ErrorCode != 0 {
	//		switch data.Config.ErrorCode {
	//		case 17, 19:
	//			// Handle the insufficient quota error.
	//			ai_convert.SetAIStatusQuotaExhausted(ctx)
	//		case 4, 18, 336501, 336502, 336503, 336504, 336505, 336507:
	//			// Handle the rate limit error.
	//			ai_convert.SetAIStatusExceeded(ctx)
	//		case 13, 14, 100, 110, 111:
	//			// Handle the invalid token error.
	//			ai_convert.SetAIStatusInvalid(ctx)
	//		default:
	//			ai_convert.SetAIStatusInvalidRequest(ctx)
	//		}
	//	} else {
	//		usage := data.Config.Usage
	//		ai_convert.SetAIStatusNormal(ctx)
	//		ai_convert.SetAIModelInputToken(ctx, usage.PromptTokens)
	//		ai_convert.SetAIModelOutputToken(ctx, usage.CompletionTokens)
	//		ai_convert.SetAIModelTotalToken(ctx, usage.TotalTokens)
	//	}
	case 400:
		// Handle the bad request error.
		ai_convert.SetAIStatusInvalidRequest(ctx)
	case 403:
		// Handle the invalid token error.
		ai_convert.SetAIStatusInvalid(ctx)
	default:
		ai_convert.SetAIStatusInvalidRequest(ctx)
	}
}
