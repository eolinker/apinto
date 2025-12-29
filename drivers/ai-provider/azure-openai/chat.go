package azure_openai

import (
	"encoding/json"
	"fmt"
	ai_convert "github.com/eolinker/apinto/ai-convert"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
	"strings"
)

func init() {
	driverCreate.Set(ai_convert.ModelTypeChat, func(conf *Config) (ai_convert.IConverterDriver, error) {
		handler, err := ai_convert.NewOpenAIChat(provider, conf.APIKey, conf.BaseUrl, 0, nil, errorCallback)
		if err != nil {
			return nil, err
		}
		return &Chat{
			apiVersion:       conf.APIVersion,
			IConverterDriver: handler,
		}, nil
	})
}

type Chat struct {
	apiVersion string
	ai_convert.IConverterDriver
}

func (c *Chat) RequestConvert(ctx eocontext.EoContext, extender map[string]interface{}) error {
	if c.IConverterDriver == nil {
		return fmt.Errorf("handler is not initialized")
	}
	err := c.IConverterDriver.RequestConvert(ctx, extender)
	if err != nil {
		return err
	}
	httpContext, err := http_service.Assert(ctx)
	if err != nil {
		return err
	}
	httpContext.Proxy().URI().SetPath(fmt.Sprintf("/openai/deployments/%s/%s", ai_convert.GetAIModel(ctx), strings.TrimPrefix(httpContext.Proxy().URI().Path(), "/v1")))
	httpContext.Proxy().URI().SetQuery("api-version", c.apiVersion)
	return nil
}

func (c *Chat) ResponseConvert(ctx eocontext.EoContext) error {
	if c.IConverterDriver != nil {
		return c.IConverterDriver.ResponseConvert(ctx)
	}
	return fmt.Errorf("handler is not initialized")
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
