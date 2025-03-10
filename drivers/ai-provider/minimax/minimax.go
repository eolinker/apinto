package minimax

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/eolinker/eosc/log"

	ai_convert "github.com/eolinker/apinto/ai-convert"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

func init() {
	ai_convert.RegisterConverterCreateFunc("minimax", Create)
}

type Config struct {
	APIKey  string `json:"minimax_api_key"`
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
