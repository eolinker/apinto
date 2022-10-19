package redis

import (
	"github.com/eolinker/eosc"
	"reflect"
)

type Driver struct {
}

func (d *Driver) Check(v interface{}, workers map[eosc.RequireId]eosc.IWorker) error {

	_, err := checkConfig(v)
	return err
}
func checkConfig(v interface{}) (*Config, error) {
	cfg, ok := v.(*Config)
	if !ok {
		return nil, eosc.ErrorConfigIsNil
	}
	return cfg, nil
}
func (d *Driver) ConfigType() reflect.Type {
	return configType
}

func (d *Driver) Create(id, name string, v interface{}, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	if err := d.Check(v, workers); err != nil {
		return nil, err
	}
	w := &Worker{
		ICache: &Empty{},
		config: nil,
		client: nil,
		id:     id,
		name:   name,
	}
	err := w.Reset(v, workers)
	if err != nil {
		return nil, err
	}
	return w, nil
}
