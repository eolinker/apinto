package visit_strategy

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

	_, err := strategy.ParseFilter(conf.Filters)
	if err != nil {
		return err
	}

	_, err = strategy.ParseFilter(conf.Rule.InfluenceSphere)
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

	lg := &Visit{
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
