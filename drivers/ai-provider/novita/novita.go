package novita

import (
	"encoding/json"
	"fmt"
	"net/url"

	ai_convert "github.com/eolinker/apinto/ai-convert"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

func init() {
	ai_convert.RegisterConverterCreateFunc("novita", Create)
}

type Config struct {
	APIKey  string `json:"api_key"`
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

	return ai_convert.NewOpenAIConvert(conf.APIKey, conf.BaseUrl, 0, errorCallback)
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
