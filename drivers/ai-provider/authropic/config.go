package anthropic

import (
	"fmt"
	"net/url"
)

const defaultVersion = "2023-06-01"

type Config struct {
	APIKey string `json:"api_key"`
	Base   string `json:"base_url"`
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
	return nil
}
