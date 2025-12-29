package lm_studio

import (
	"fmt"
	"net/url"
)

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
