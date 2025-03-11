package lm_studio

import (
	"encoding/json"
	"fmt"
	"net/url"

	ai_convert "github.com/eolinker/apinto/ai-convert"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

func init() {
	ai_convert.RegisterConverterCreateFunc("lm-studio", Create)
}

type Config struct {
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
	if conf.BaseUrl == "" {
		return fmt.Errorf("base url is required")
	}
	u, err := url.Parse(conf.BaseUrl)
	if err != nil {
		// Return an error if the Base URL cannot be parsed.
		return fmt.Errorf("base url is invalid")
	}
	// Ensure the parsed URL contains both a scheme and a host.
	if u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("base url is invalid")
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

	return ai_convert.NewOpenAIConvert("", conf.BaseUrl, 0, nil, errorCallback)
}

func errorCallback(ctx http_service.IHttpContext, body []byte) {

	switch ctx.Response().StatusCode() {
	case 400:
		// Handle the bad request error.
		ai_convert.SetAIStatusInvalidRequest(ctx)
	case 401:
		// Handle the key error.
		ai_convert.SetAIStatusInvalid(ctx)
	default:
		ai_convert.SetAIStatusInvalidRequest(ctx)
	}
}
