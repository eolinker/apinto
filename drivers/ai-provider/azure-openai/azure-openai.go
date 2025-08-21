package azure_openai

import (
	"encoding/json"
	"fmt"
	"github.com/eolinker/eosc/eocontext"
	"net/url"

	"github.com/eolinker/eosc/log"

	ai_convert "github.com/eolinker/apinto/ai-convert"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

func init() {
	ai_convert.RegisterConverterCreateFunc("azure_openai", Create)
}

type Config struct {
	APIKey     string `json:"api_key"`
	BaseUrl    string `json:"base_url"`
	APIVersion string `json:"api_version"`
}

// checkConfig validates the provided configuration.
// It ensures the required fields are set and checks the validity of the Base URL if provided.
//
// Parameters:
//   - v: An interface{} expected to be a pointer to a Config struct.
//
// Returns:
//   - *Config: The validated configuration cast to *Config.
//   - error: An error if the validation fails, or nil if it succeeds.
func checkConfig(conf *Config) error {
	// Check if the APIKey is provided. It is a required field.
	if conf.APIKey == "" {
		return fmt.Errorf("api_key is required")
	}
	if conf.BaseUrl != "" {
		u, err := url.Parse(conf.BaseUrl)
		if err != nil {
			// Return an error if the Base URL cannot be parsed.
			return fmt.Errorf("base url is invalid")
		}
		// Ensure the parsed URL contains both a scheme and a host.
		if u.Scheme == "" || u.Host == "" {
			return fmt.Errorf("base url is invalid")
		}
	}
	if conf.APIVersion == "" {
		return fmt.Errorf("api_version is required")
	}
	return nil
}

func NewChat(apiVersion string, handler ai_convert.IConverter) (ai_convert.IConverter, error) {
	return &Chat{apiVersion: apiVersion, handler: handler}, nil
}

type Chat struct {
	apiVersion string
	handler    ai_convert.IConverter
}

func (c *Chat) RequestConvert(ctx eocontext.EoContext, extender map[string]interface{}) error {
	if c.handler == nil {
		return fmt.Errorf("handler is not initialized")
	}
	httpContext, err := http_service.Assert(ctx)
	if err != nil {
		return err
	}
	httpContext.Proxy().URI().SetQuery("api-version", c.apiVersion)
	return c.handler.RequestConvert(httpContext, extender)
}

func (c *Chat) ResponseConvert(ctx eocontext.EoContext) error {
	if c.handler != nil {
		return c.handler.ResponseConvert(ctx)
	}
	return fmt.Errorf("handler is not initialized")
}

func Create(cfg string) (ai_convert.IConverter, error) {
	var conf Config
	err := json.Unmarshal([]byte(cfg), &conf)
	if err != nil {
		return nil, err
	}
	err = checkConfig(&conf)
	if err != nil {
		return nil, err
	}

	handler, err := ai_convert.NewOpenAIConvert(conf.APIKey, conf.BaseUrl, 0, nil, errorCallback)
	if err != nil {
		return nil, err
	}
	return NewChat(conf.APIVersion, handler)
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
