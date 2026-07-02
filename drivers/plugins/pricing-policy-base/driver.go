package pricing_policy_base

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

func Create(id, name string, v *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	w := &executor{
		WorkerBase: drivers.Worker(id, name),
	}
	err := w.reset(v)
	if err != nil {
		return nil, err
	}
	return w, nil
}
