package openAI

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/eolinker/eosc/log"

	http_service "github.com/eolinker/eosc/eocontext/http-context"

	ai_convert "github.com/eolinker/apinto/ai-convert"
)

func init() {
	ai_convert.RegisterConverterCreateFunc("openai", Create)
}

// Config represents the configuration for OpenAI API.
// It includes the necessary fields for authentication and base URL configuration.
type Config struct {
	APIKey       string `json:"openai_api_key"`      // APIKey is the authentication key for accessing OpenAI API.
	Organization string `json:"openai_organization"` // Organization specifies the associated organization ID (optional).
	Base         string `json:"openai_api_base"`     // Base is the base URL for OpenAI API. It can be customized if needed.
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

	// Validate the Base URL if it is provided.
	if conf.Base != "" {
		u, err := url.Parse(conf.Base)
		if err != nil {
			// Return an error if the Base URL cannot be parsed.
			return fmt.Errorf("base url is invalid")
		}
		// Ensure the parsed URL contains both a scheme and a host.
		if u.Scheme == "" || u.Host == "" {
			return fmt.Errorf("base url is invalid")
		}
	}

	// Return the validated configuration.
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

	return ai_convert.NewOpenAIConvert(conf.APIKey, conf.Base, 0, nil, errorCallback)
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
