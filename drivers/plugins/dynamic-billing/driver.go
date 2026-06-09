package dynamic_billing

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

// Create 实例化 worker
func Create(id, name string, v *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	w := &executor{
		WorkerBase: drivers.Worker(id, name),
	}
	err := w.reset(v, workers)
	return w, err
}
