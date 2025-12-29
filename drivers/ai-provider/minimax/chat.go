package minimax

import (
	"encoding/json"
	ai_convert "github.com/eolinker/apinto/ai-convert"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
)

func init() {
	driverCreate.Set(ai_convert.ModelTypeChat, func(c *Config) (ai_convert.IConverterDriver, error) {
		return ai_convert.NewOpenAIChat(provider, c.APIKey, c.BaseUrl, 0, checkError, errorCallback)
	})
}

func checkError(ctx http_service.IHttpContext, body []byte) bool {
	if ctx.Response().StatusCode() != 200 {
		return false
	}
	var data Response
	err := json.Unmarshal(body, &data)
	if err != nil {
		log.Errorf("Failed to unmarshal response body: %v", err)
		return false
	}
	return data.BaseResp.StatusCode == 0
}

func errorCallback(ctx http_service.IHttpContext, body []byte) {
	var data Response
	err := json.Unmarshal(body, &data)
	if err != nil {
		log.Errorf("Failed to unmarshal response body: %v", err)
		return
	}
	switch data.BaseResp.StatusCode {
	case 2013: // 输入格式信息不正常
		// Handle the bad request error.
		ai_convert.SetAIStatusInvalidRequest(ctx)
	case 1008:
		// Handle the balance is insufficient.
		ai_convert.SetAIStatusQuotaExhausted(ctx)
	case 1002, 1039: // 触发RPM限流 || 触发TPM限流
		// Handle exceed
		ai_convert.SetAIStatusExceeded(ctx)
	case 1004:
		// Handle authentication failure
		ai_convert.SetAIStatusInvalid(ctx)
	}
}
