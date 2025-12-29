package openAI

import "fmt"

// Config represents the configuration for OpenAI API.
// It includes the necessary fields for authentication and base URL configuration.
type Config struct {
	APIKey  string `json:"api_key"` // APIKey is the authentication key for accessing OpenAI API.
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

	return nil
}
