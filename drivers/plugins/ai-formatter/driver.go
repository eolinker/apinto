package ai_formatter

import (
	"strings"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

func Create(id, name string, v *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	_, err := checkConfig(v)
	if err != nil {
		return nil, err
	}
	provider := strings.Split(v.Provider, "@")

	w := &executor{
		WorkerBase: drivers.Worker(id, name),
		model:      v.Model,
		modelCfg:   v.Config,
		provider:   provider[0],
	}
	if err != nil {
		return nil, err
	}
	return w, err
}
