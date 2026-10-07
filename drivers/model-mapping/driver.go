package model_mapping

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

// CheckConfig 校验配置与依赖 worker
func CheckConfig(cfg *Config, workers map[eosc.RequireId]eosc.IWorker) error {
	_, err := Check(cfg)
	return err
}

// Create 实例化策略 worker
func Create(id, name string, v *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	if err := CheckConfig(v, workers); err != nil {
		return nil, err
	}

	lg := &executor{
		WorkerBase: drivers.Worker(id, name),
	}

	err := lg.Reset(v, workers)
	if err != nil {
		return nil, err
	}

	if controller != nil {
		controller.Store(id)
	}
	return lg, nil
}
