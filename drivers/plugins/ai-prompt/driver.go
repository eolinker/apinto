package ai_prompt

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

func Create(id, name string, v *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	w := &executor{
		WorkerBase: drivers.Worker(id, name),
	}
	err := w.Reset(v, workers)
	return w, err
}
