package quota_limiting_strategy

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

func Check(cfg *Config, workers map[eosc.RequireId]eosc.IWorker) error {
	return checkConfig(cfg)
}

func Create(id, name string, v *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	if err := Check(v, workers); err != nil {
		return nil, err
	}
	
	q := &executor{
		WorkerBase: drivers.Worker(id, name),
	}
	
	err := q.Reset(v, workers)
	if err != nil {
		return nil, err
	}
	
	if controller != nil {
		controller.Store(id)
	}
	return q, nil
}
