package rewrite

import (
	"reflect"

	"github.com/eolinker/eosc"
)

type Driver struct {
	profession string
	name       string
	label      string
	desc       string
	workers    eosc.IWorkers
	configType reflect.Type
}

func (d *Driver) Check(v interface{}, workers map[eosc.RequireId]interface{}) error {
	_, err := d.check(v)
	if err != nil {
		return err
	}
	return nil
}
func (d *Driver) check(v interface{}) (*Config, error) {
	conf, ok := v.(*Config)
	if !ok {
		return nil, eosc.ErrorConfigFieldUnknown
	}
	return conf, nil
}
func (d *Driver) ConfigType() reflect.Type {
	return d.configType
}

func (d *Driver) Create(id, name string, v interface{}, workers map[eosc.RequireId]interface{}) (eosc.IWorker, error) {
	conf, err := d.check(v)
	if err != nil {
		return nil, err
	}

	rw := &Rewrite{
		Driver: d,
		id:     id,
		name:   name,
		url:    conf.ReWriteUrl,
	}

	return rw, nil
}
