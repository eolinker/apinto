package zhinao

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
	switch ctx.Response().StatusCode() {
	case 400:
		// Handle the bad request error.
		ai_convert.SetAIStatusInvalidRequest(ctx)
	case 429:
		// Handle exceed
		ai_convert.SetAIStatusExceeded(ctx)
	case 401:
		var data ai_convert.Response
		err := json.Unmarshal(body, &data)
		if err != nil {
			log.Errorf("unmarshal body error: %v, body: %s", err, body)
			return
		}
		if data.Error.Code == "1004" {
			// Handle the balance is insufficient.
			ai_convert.SetAIStatusQuotaExhausted(ctx)
		} else if data.Error.Code == "1006" {
			// 日限额
			ai_convert.SetAIStatusExceeded(ctx)
		} else {
			// Handle authentication failure
			ai_convert.SetAIStatusInvalid(ctx)
		}
	}
}
