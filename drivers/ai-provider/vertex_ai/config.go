package vertex_ai

import (
	"encoding/base64"
	"fmt"

	"golang.org/x/oauth2/google"

	"github.com/eolinker/eosc"
)

type Config struct {
	ProjectID         string `json:"vertex_project_id"`
	Location          string `json:"vertex_location"`
	ServiceAccountKey string `json:"vertex_service_account_key"`
	Base              string `json:"vertex_api_base"`
}

func checkConfig(v interface{}) (*Config, error) {
	conf, ok := v.(*Config)
	if !ok {
		return nil, eosc.ErrorConfigType
	}
	if conf.ProjectID == "" {
		return nil, fmt.Errorf("project_id is required")
	}
	if conf.Location == "" {
		return nil, fmt.Errorf("location is required")
	}
	serviceAccountKey, err := base64.RawStdEncoding.DecodeString(conf.ServiceAccountKey)
	if err != nil {
		return nil, err
	}
	_, err = google.JWTConfigFromJSON(serviceAccountKey)
	if err != nil {
		return nil, err
	}
	return conf, nil
}
