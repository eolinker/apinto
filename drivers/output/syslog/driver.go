package syslog

import (
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/formatter"
	"github.com/eolinker/eosc/utils/schema"
	"reflect"
)

type Driver struct {
	configType reflect.Type
}

func (d *Driver) ConfigType() reflect.Type {
	return d.configType
}

func (d *Driver) Render() interface{} {
	render, err := schema.Generate(reflect.TypeOf((*Config)(nil)), nil)
	if err != nil {
		return nil
	}
	return render
}

func (d *Driver) check(v interface{}) (*Config, error) {
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

func (d *Driver) Create(id, name string, v interface{}, workers map[eosc.RequireId]interface{}) (eosc.IWorker, error) {
	cfg, err := d.check(v)
	if err != nil {
		return nil, err
	}
	worker, err := CreateTransporter(cfg)
	if err != nil {
		return nil, err
	}
	// 新建formatter
	factory, has := formatter.GetFormatterFactory(cfg.Type)
	if !has {
		return nil, errFormatterType
	}
	format, err := factory.Create(cfg.Formatter)
	if err != nil {
		return nil, err
	}
	worker.formatter = format
	worker.id = id
	worker.Driver = d
	return worker, nil
}
