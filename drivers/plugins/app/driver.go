package app

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

func Create(id, name string, v *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	return &App{
		WorkerBase: drivers.Worker(id, name),
	}, nil
}
