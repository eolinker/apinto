package vertex_ai

import (
	"fmt"

	"github.com/eolinker/eosc"
)

type Config struct {
	ProjectID         string `json:"vertex_project_id"`
	Location          string `json:"vertex_location"`
	ServiceAccountKey string `json:"vertex_service_account_key"`
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
	return conf, nil
}
