package openAI

import (
	"fmt"
	"net/url"

	"github.com/eolinker/eosc"
)

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
func checkConfig(v interface{}) (*Config, error) {
	// Attempt to cast the input to a *Config type.
	conf, ok := v.(*Config)
	if !ok {
		// Return an error if the type is incorrect.
		return nil, eosc.ErrorConfigType
	}

	// Check if the APIKey is provided. It is a required field.
	if conf.APIKey == "" {
		return nil, fmt.Errorf("api_key is required")
	}

	// Validate the Base URL if it is provided.
	if conf.Base != "" {
		u, err := url.Parse(conf.Base)
		if err != nil {
			// Return an error if the Base URL cannot be parsed.
			return nil, fmt.Errorf("base url is invalid")
		}
		// Ensure the parsed URL contains both a scheme and a host.
		if u.Scheme == "" || u.Host == "" {
			return nil, fmt.Errorf("base url is invalid")
		}
	}

	// Return the validated configuration.
	return conf, nil
}
