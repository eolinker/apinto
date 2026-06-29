package openAI

import (
	"encoding/json"
	ai_convert "github.com/eolinker/apinto/ai-convert"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
)

func init() {
	driverCreate.Set(ai_convert.ModelTypeOpenAIChat, func(mt ai_convert.ModelType, c *Config) (ai_convert.IConverterDriver, error) {
		return ai_convert.NewOpenAIChat(provider, c.APIKey, c.Base, mt, 0, nil, errorCallback)
	})
	driverCreate.Set(ai_convert.ModelTypeChat, func(mt ai_convert.ModelType, c *Config) (ai_convert.IConverterDriver, error) {
		return ai_convert.NewOpenAIChat(provider, c.APIKey, c.Base, mt, 0, nil, errorCallback)
	})
}

func errorCallback(ctx http_service.IHttpContext, body []byte) {
	var resp ai_convert.Response
	err := json.Unmarshal(body, &resp)
	if err != nil {
		log.Errorf("unmarshal body error: %v, body: %s", err, string(body))
		return
	}
	switch ctx.Response().StatusCode() {

	case 400:
		// Handle the bad request error.
		ai_convert.SetAIStatusInvalidRequest(ctx)
	case 429:
		switch resp.Error.Type {
		case "insufficient_quota":
			// Handle the insufficient quota error.
			ai_convert.SetAIStatusQuotaExhausted(ctx)
		case "rate_limit_error":
			// Handle the rate limit error.
			ai_convert.SetAIStatusExceeded(ctx)
		}
	case 401:
		// 过期和无效的API密钥
		ai_convert.SetAIStatusInvalid(ctx)
	}
}
