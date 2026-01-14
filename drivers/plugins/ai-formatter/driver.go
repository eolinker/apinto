package ai_formatter

import (
	ai_convert "github.com/eolinker/apinto/ai-convert"
	"strings"
	
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

func Create(id, name string, v *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	conf, err := checkConfig(v)
	if err != nil {
		return nil, err
	}
	provider := strings.Split(v.Provider, "@")
	
	w := &executor{
		WorkerBase: drivers.Worker(id, name),
		model:      conf.Model,
		modelType:  ai_convert.ModelType(conf.ModelType),
		modelCfg:   conf.Config,
		provider:   provider[0],
	}
	if err != nil {
		return nil, err
	}
	return w, err
}
