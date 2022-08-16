package rate_limiting

import (
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/utils/schema"
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
if err != nil {
return err
}
return nil
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
	return conf, nil
}

func (d *Driver) Create(id, name string, v interface{}, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	conf, err := d.check(v)
	if err != nil {
		return nil, err
	}
	l := &RateLimiting{
		Driver:           d,
		id:               id,
		rateInfo:         CreateRateInfo(conf),
		hideClientHeader: conf.HideClientHeader,
		responseType:     conf.ResponseType,
	}
	return l, nil
}
