package vertex_ai

import (
	"encoding/base64"
	"fmt"
	"golang.org/x/oauth2/google"
	"net/url"
	"strings"
)

type Config struct {
	ProjectID         string `json:"vertex_project_id"`
	Location          string `json:"vertex_location"`
	ServiceAccountKey string `json:"vertex_service_account_key"`
	Base              string `json:"vertex_api_base"`
}

func checkConfig(conf *Config) error {
	// Check if the APIKey is provided. It is a required field.
	if conf.ProjectID == "" {
		return fmt.Errorf("project_id is required")
	}
	if conf.Location == "" {
		return fmt.Errorf("location is required")
	}
	if conf.ServiceAccountKey == "" {
		return fmt.Errorf("service_account_key is required")
	}
	serviceAccountKey, err := base64.StdEncoding.DecodeString(conf.ServiceAccountKey)
	_, err = google.JWTConfigFromJSON(serviceAccountKey)
	if err != nil {
		return err
	}
	if conf.Base != "" {
		tmpBase := ""
		tmpBase = strings.ReplaceAll(conf.Base, "{LOCATION}", conf.Location)
		tmpBase = strings.ReplaceAll(tmpBase, "{PROJECT_ID}", conf.ProjectID)
		u, err := url.Parse(tmpBase)
		if err != nil {
			// Return an error if the BaseUrl URL cannot be parsed.
			return fmt.Errorf("base url is invalid")
		}
		// Ensure the parsed URL contains both a scheme and a host.
		if u.Scheme == "" || u.Host == "" {
			return fmt.Errorf("base url is invalid")
		}
	}

	return nil
}
