package openAI

import (
	"fmt"
	"net/url"
)

// Config represents the configuration for OpenAI API.
// It includes the necessary fields for authentication and base URL configuration.
type Config struct {
	APIKey  string `json:"api_key"`  // APIKey is the authentication key for accessing OpenAI API.
	BaseUrl string `json:"base_url"` // BaseUrl is the base URL for OpenAI API. It can be customized if needed.
}

// checkConfig validates the provided configuration.
// It ensures the required fields are set and checks the validity of the BaseUrl URL if provided.
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

	// Validate the BaseUrl URL if it is provided.
	if conf.BaseUrl != "" {
		u, err := url.Parse(conf.BaseUrl)
		if err != nil {
			// Return an error if the BaseUrl URL cannot be parsed.
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
