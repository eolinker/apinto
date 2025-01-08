package ai_key

import (
	"fmt"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

type Config struct {
	Expired  int64  `json:"expired"`
	Config   string `json:"config"`
	Provider string `json:"provider"`
	Priority int    `json:"priority"`
	Disabled bool   `json:"disabled"`
}

func checkConfig(v interface{}) (*Config, error) {
	conf, ok := v.(*Config)
	if !ok {
		return nil, eosc.ErrorConfigType
	}

	if conf.Provider == "" {
		return nil, fmt.Errorf("provider is required")
	}

	return conf, nil

}
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
