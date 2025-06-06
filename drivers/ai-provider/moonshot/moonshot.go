package moonshot

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/eolinker/eosc/log"

	ai_convert "github.com/eolinker/apinto/ai-convert"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

func init() {
	ai_convert.RegisterConverterCreateFunc("moonshot", Create)
}

type Config struct {
	APIKey  string `json:"moonshot_api_key"`
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

	return ai_convert.NewOpenAIConvert(conf.APIKey, conf.BaseUrl, 0, nil, errorCallback)
}

func errorCallback(ctx http_service.IHttpContext, body []byte) {
	// HTTP Status Codes for Moonshot API
	// Status Code | Type                | Error Message
	// ------------|---------------------|-------------------------------------
	// 200         | Success             | Request was successful.
	// 400         | Client Error        | Invalid request parameters (invalid_request_error).
	// 401         | Authentication Error | Invalid API key (invalid_key).
	// 403         | Forbidden           | Access denied (forbidden_error).
	// 404         | Not Found           | Resource not found (not_found_error).
	// 429         | Rate Limit Exceeded | Too many requests (rate_limit_error).
	// 500         | Server Error        | Internal server error (server_error).
	// 503         | Service Unavailable  | Service is temporarily unavailable (service_unavailable).
	switch ctx.Response().StatusCode() {

	case 400:
		// Handle the bad request error.
		ai_convert.SetAIStatusInvalidRequest(ctx)
	case 401:
		// 过期和无效的API密钥
		ai_convert.SetAIStatusInvalid(ctx)
	case 429:
		var data ai_convert.Response
		err := json.Unmarshal(body, &data)
		if err != nil {
			log.Errorf("unmarshal body error: %v, body: %s", err, body)
			return
		}
		switch data.Error.Type {
		case "exceeded_current_quota_error":
			// Handle the insufficient quota error.
			ai_convert.SetAIStatusQuotaExhausted(ctx)
		case "engine_overloaded_error", "rate_limit_reached_error":
			// Handle the rate limit error.
			ai_convert.SetAIStatusExceeded(ctx)
		}
	default:
		ai_convert.SetAIStatusInvalidRequest(ctx)
	}
}
