package cors

import (
	"github.com/eolinker/eosc"
	"reflect"
)

type Driver struct {
	profession string
	name       string
	label      string
	desc       string
	configType reflect.Type
}

func (d *Driver) Check(v interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	_, err := d.check(v)
	return err
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

func (d *Driver) ConfigType() reflect.Type {
	return d.configType
}

func (d *Driver) Create(id, name string, v interface{}, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {

	conf, err := d.check(v)
	if err != nil {
		return nil, err
	}
	c := &CorsFilter{
		Driver:           d,
		id:               id,
		option:           conf.genOptionHandler(),
		originChecker:    NewChecker(conf.AllowOrigins, "Access-Control-Allow-Origin"),
		methodChecker:    NewChecker(conf.AllowMethods, "Access-Control-Allow-Methods"),
		headerChecker:    NewChecker(conf.AllowHeaders, "Access-Control-Allow-Headers"),
		exposeChecker:    NewChecker(conf.ExposeHeaders, "Access-Control-Expose-Headers"),
		allowCredentials: conf.AllowCredentials,
	}
	return c, nil
}
