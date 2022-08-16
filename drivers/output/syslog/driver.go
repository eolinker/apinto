package syslog

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

func check(v interface{}) (*Config, error) {
	conf, ok := v.(*Config)
	if !ok {
		return nil, eosc.ErrorConfigFieldUnknown
	}
	err := conf.doCheck()
	if err != nil {
		return nil, err
	}
	return conf, nil

}

func (d *Driver) Create(id, name string, v interface{}, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	cfg, err := check(v)

	if err != nil {
		return nil, err
	}

	return &Output{
		id:     id,
		name:   name,
		config: cfg,
		writer: nil,
	}, nil

}
