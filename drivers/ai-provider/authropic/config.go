package anthropic

import (
	"fmt"
	"net/url"
)

const defaultVersion = "2023-06-01"

type Config struct {
	APIKey  string `json:"anthropic_api_key"`
	Base    string `json:"anthropic_api_url"`
	Version string `json:"anthropic_api_version"`
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
