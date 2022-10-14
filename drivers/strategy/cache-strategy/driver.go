package cache_strategy

import (
	"fmt"
	"github.com/eolinker/apinto/strategy"
	"github.com/eolinker/eosc"
	"reflect"
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

type driver struct {
}

func (d *driver) Check(v interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	cfg, ok := v.(*Config)
	if !ok {
		return eosc.ErrorConfigIsNil
	}

	return checkConfig(cfg)
}

func (d *driver) ConfigType() reflect.Type {
	return configType
}

func (d *driver) Create(id, name string, v interface{}, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	if err := d.Check(v, workers); err != nil {
		return nil, err
	}

	lg := &CacheValidTime{
		id:   id,
		name: name,
	}

	err := lg.Reset(v, workers)
	if err != nil {
		return nil, err
	}

	controller.Store(id)
	return lg, nil
}
