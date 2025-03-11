package openAI

import (
	"encoding/json"
	"fmt"

	"github.com/eolinker/eosc/log"

	http_service "github.com/eolinker/eosc/eocontext/http-context"

	ai_convert "github.com/eolinker/apinto/ai-convert"
)

func init() {
	ai_convert.RegisterConverterCreateFunc("volcengine_mass", Create)
}

// Config represents the configuration for OpenAI API.
// It includes the necessary fields for authentication and base URL configuration.
type Config struct {
	APIKey  string `json:"api_key"` // APIKey is the authentication key for accessing OpenAI API.
	BaseUrl string `json:"base_url"`
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

	return nil
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
	if conf.BaseUrl == "" {
		conf.BaseUrl = "https://ark.cn-beijing.volces.com/api/v3"
	}
	return ai_convert.NewOpenAIConvert(conf.APIKey, conf.BaseUrl, 0, nil, errorCallback)
}

func errorCallback(ctx http_service.IHttpContext, body []byte) {
	var resp ai_convert.Response
	err := json.Unmarshal(body, &resp)
	if err != nil {
		log.Errorf("unmarshal body error: %v, body: %s", err, string(body))
		return
	}
	switch ctx.Response().StatusCode() {
	case 400, 404:
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
