package cors

import (
	"reflect"

	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/drivers"
)

type Driver struct {
	profession string
	name       string
	label      string
	desc       string
	configType reflect.Type
}

func Check(v *Config, workers map[eosc.RequireId]eosc.IWorker) error {

	return v.doCheck()
}

func check(v interface{}) (*Config, error) {
	conf, ok := v.(*Config)
	if !ok {
		return nil, eosc.ErrorConfigType
	}
	err := conf.doCheck()
	if err != nil {
		return nil, err
	}
	return conf, nil
}

func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {

	err := Check(conf, workers)
	if err != nil {
		return nil, err
	}
	c := &CorsFilter{
		WorkerBase:       drivers.Worker(id, name),
		option:           conf.genOptionHandler(),
		originChecker:    NewChecker(conf.AllowOrigins, "Access-Control-Allow-Origin"),
		methodChecker:    NewChecker(conf.AllowMethods, "Access-Control-Allow-Methods"),
		headerChecker:    NewChecker(conf.AllowHeaders, "Access-Control-Allow-Headers"),
		exposeChecker:    NewChecker(conf.ExposeHeaders, "Access-Control-Expose-Headers"),
		allowCredentials: conf.AllowCredentials,
	}
	return c, nil
}
