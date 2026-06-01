package baichuan

import (
	"encoding/json"
	ai_convert "github.com/eolinker/apinto/ai-convert"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
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
	case 429:
		var data ai_convert.Response
		err := json.Unmarshal(body, &data)
		if err != nil {
			log.Errorf("unmarshal response error: %v, body: %s", err, body)
			return
		}
		if data.Error.Code == "insufficient_quota" {
			// Handle the balance is insufficient.
			ai_convert.SetAIStatusQuotaExhausted(ctx)
		} else {
			// Handle exceed
			ai_convert.SetAIStatusExceeded(ctx)
		}
	case 401:
		// Handle authentication failure
		ai_convert.SetAIStatusInvalid(ctx)
	}
}
