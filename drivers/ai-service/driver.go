package ai_service

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

// Create 创建实例
func Create(id, name string, v *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {

	w := &executor{
		WorkerBase: drivers.Worker(id, name),
		title:      v.Title,
	}

	err := w.Reset(v, workers)
	if err != nil {
		return nil, err
	}

	return w, nil
}
