package ai_provider

import (
	"context"
	"fmt"

	"github.com/eolinker/apinto/drivers"

	"github.com/eolinker/eosc"
)

type Config struct {
	Provider    string `json:"provider"`
	Model       string `json:"model"`
	ModelConfig string `json:"model_config"`
	Priority    int    `json:"priority"`
	Keys        []*Key `json:"keys"`
}

type Key struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Expired  int64  `json:"expired"`
	Config   string `json:"config"`
	Disabled bool   `json:"disabled"`
}

// Create 创建驱动实例
func Create(id, name string, v *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	cfg, err := checkConfig(v)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	w := &executor{
		WorkerBase: drivers.Worker(id, name),
		ctx:        ctx,
		cancel:     cancel,
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

	if len(conf.Keys) == 0 {
		return nil, fmt.Errorf("keys is required")
	}
	return conf, nil
}
