package limiting_stragety

import (
	"github.com/eolinker/apinto/strategy"
	"github.com/eolinker/eosc"
	"reflect"
)

type driver struct {
	configType reflect.Type
}

func (d *driver) Check(v interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	cfg, ok := v.(*Config)
	if !ok {
		return eosc.ErrorConfigIsNil
	}
	_, err := strategy.ParseFilter(cfg.Filters)
	if err != nil {
		return err
	}

	return nil
}

func (d *driver) ConfigType() reflect.Type {
	return d.configType
}

func (d *driver) Create(id, name string, v interface{}, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	return &Limiting{
		id: id,
	}, nil
}
