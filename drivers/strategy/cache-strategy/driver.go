package cache_strategy

import (
	"fmt"

	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/strategy"
)

func checkConfig(conf *Config) error {
	if conf.Priority > 999 || conf.Priority < 1 {
		return fmt.Errorf("priority value %d not allow ", conf.Priority)
	}

	if conf.ValidTime < 1 {
		return fmt.Errorf("validTime value %d not allow ", conf.ValidTime)
	}

	_, err := strategy.ParseFilter(conf.Filters)
	if err != nil {
		return err
	}

	return nil
}

func Check(v *Config, workers map[eosc.RequireId]eosc.IWorker) error {

	return checkConfig(v)
}

func Create(id, name string, v *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	if err := Check(v, workers); err != nil {
		return nil, err
	}

	lg := &CacheValidTime{
		WorkerBase: drivers.Worker(id, name),
	}

	err := lg.Reset(v, workers)
	if err != nil {
		return nil, err
	}

	controller.Store(id)
	return lg, nil
}
