package redis

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

func Create(id, name string, v *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {

	w := &Worker{
		WorkerBase: drivers.Worker(id, name),
		ICache:     &Empty{},
		IVectors:   &Empty{},
		config:     nil,
		client:     nil,
	}
	err := w.Reset(v, workers)
	if err != nil {
		return nil, err
	}
	return w, nil
}
