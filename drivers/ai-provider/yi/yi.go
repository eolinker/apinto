package yi

import (
	"encoding/json"
	"fmt"
	"net/url"

	ai_convert "github.com/eolinker/apinto/ai-convert"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

func init() {
	ai_convert.RegisterConverterCreateFunc("yi", Create)
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

	return ai_convert.NewOpenAIConvert(conf.APIKey, conf.BaseUrl, 0, nil, errorCallback)
}

func errorCallback(ctx http_service.IHttpContext, body []byte) {
	/*
		错误码对照表
		| HTTP 返回码 | 错误代码     | 原因                                   | 解决方案                             |
		|------------|------------|----------------------------------------|--------------------------------------|
		| 400        | Bad request | 模型的输入+输出（max_tokens）超过了模型的最大上下文。 | 减少模型的输入，或将 max_tokens 参数值设置更小。 |
		|            |            | 输入格式错误。                           | 检查输入格式，确保正确。例如，模型名必须全小写，yi-lightning。 |
		| 401        | Authentication Error | API Key缺失或无效。             | 请确保你的 API Key 有效。               |
		| 404        | Not found   | 无效的 Endpoint URL 或模型名。           | 确保使用正确的 Endpoint URL 或模型名。   |
		| 429        | Too Many Requests | 在短时间内发出的请求太多。         | 控制请求速率。                       |
		| 500        | Internal Server Error | 服务端内部错误。           | 请稍后重试。                         |
		| 529        | System busy | 系统繁忙，请重试。                     | 请 1 分钟后重试。                     |
	*/
	switch ctx.Response().StatusCode() {
	case 400:
		// Handle the bad request error.
		ai_convert.SetAIStatusInvalidRequest(ctx)
	case 429:
		ai_convert.SetAIStatusExceeded(ctx)
	case 401:
		// Handle the authentication error.
		ai_convert.SetAIStatusInvalid(ctx)
	}
}
