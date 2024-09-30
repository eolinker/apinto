package ai_formatter

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

func Create(id, name string, v *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	_, err := checkConfig(v)
	if err != nil {
		return nil, err
	}
	w := &executor{
		WorkerBase: drivers.Worker(id, name),
	}
	err = w.reset(v, workers)
	if err != nil {
		return nil, err
	}
	return w, err
}
