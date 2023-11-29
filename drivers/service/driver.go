package service

import (
	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/drivers"
)

// Create 创建service_http驱动的实例
func Create(id, name string, v *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {

	w := &serviceWorker{
		WorkerBase: drivers.Worker(id, name),
		Service:    Service{},
	}

	err := w.Reset(v, workers)
	if err != nil {
		return nil, err
	}

	return w, nil
}
