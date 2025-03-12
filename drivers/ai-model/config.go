package ai_model

import (
	"fmt"

	"github.com/eolinker/apinto/drivers"

	"github.com/eolinker/eosc"
)

type Config struct {
	Provider     string `json:"provider"`
	Model        string `json:"model"`
	AccessConfig string `json:"access_config"`
}

// Create 创建驱动实例
func Create(id, name string, v *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	cfg, err := checkConfig(v)
	if err != nil {
		return nil, err
	}
	w := &executor{
		WorkerBase: drivers.Worker(id, name),
	}
	err = w.reset(cfg)
	return w, err
}

func checkConfig(v interface{}) (*Config, error) {
	conf, ok := v.(*Config)
	if !ok {
		return nil, eosc.ErrorConfigType
	}

	if conf.Provider == "" {
		return nil, fmt.Errorf("provider is required")
	}

	if conf.Model == "" {
		return nil, fmt.Errorf("model is required")
	}

	return conf, nil
}
