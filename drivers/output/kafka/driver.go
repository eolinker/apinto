package kafka

import (
	"github.com/eolinker/eosc"
	"reflect"
)

type Driver struct {
	configType reflect.Type
}

func (d *Driver) ConfigType() reflect.Type {
	return d.configType
}

func (d *Driver) check(v interface{}) (*Config, error) {
	//conf, ok := v.(*Config)
	//if !ok {
	//	return nil, eosc.ErrorConfigFieldUnknown
	//}
	//err := conf.doCheck()
	//if err != nil {
	//	return nil, err
	//}
	//return conf, nil
	panic("")
}

func (d *Driver) Create(id, name string, v interface{}, workers map[eosc.RequireId]interface{}) (eosc.IWorker, error) {
	//conf, err := d.check(v)
	//if err != nil {
	//	return nil, err
	//}
	//return &Output{
	//	Driver: d,
	//	id:     id,
	//}, nil
	panic("")
}
