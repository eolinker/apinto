package anthropic

import (
	"encoding/json"
	"fmt"
	"net/url"

	http_context "github.com/eolinker/eosc/eocontext/http-context"

	ai_convert "github.com/eolinker/apinto/ai-convert"
)

const defaultVersion = "2023-06-01"

type Config struct {
	APIKey  string `json:"anthropic_api_key"`
	Base    string `json:"anthropic_api_url"`
	Version string `json:"anthropic_api_version"`
}

var name = "anthropic"

func init() {
	ai_convert.RegisterConverterCreateFunc(name, Create)
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

func checkConfig(conf *Config) error {

	if conf.APIKey == "" {
		return fmt.Errorf("api_key is required")
	}
	if conf.Base != "" {
		u, err := url.Parse(conf.Base)
		if err != nil {
			return fmt.Errorf("base url is invalid")
		}
		if u.Scheme == "" || u.Host == "" {
			return fmt.Errorf("base url is invalid")
		}
	}
	if conf.Version == "" {
		conf.Version = defaultVersion
	}
	return nil
}

func errorCallback(ctx http_context.IHttpContext, body []byte) {
	switch ctx.Response().StatusCode() {
	case 400:
		// Handle the bad request error.
		ai_convert.SetAIStatusInvalidRequest(ctx)
	case 429:
		// Handle exceed
		ai_convert.SetAIStatusExceeded(ctx)
	case 401:
		// Handle authentication failure
		ai_convert.SetAIStatusInvalid(ctx)
	}
}
