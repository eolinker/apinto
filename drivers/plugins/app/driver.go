package app

import (
	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/drivers"
)

func Create(id, name string, v *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	return &App{
		WorkerBase: drivers.Worker(id, name),
	}, nil
}
