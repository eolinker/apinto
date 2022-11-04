package params_transformer

import (
	"github.com/eolinker/apinto/drivers"
	"reflect"

	"github.com/eolinker/eosc"
)

type Driver struct {
	profession string
	name       string
	label      string
	desc       string
	configType reflect.Type
}

func Check(conf *Config, workers map[eosc.RequireId]eosc.IWorker) error {

	return conf.doCheck()
}

func check(v interface{}) (*Config, error) {

	conf, err := drivers.Assert[Config](v)
	if err != nil {
		return nil, err
	}
	err = conf.doCheck()
	if err != nil {
		return nil, err
	}

	return conf, nil
}

func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {

	err := conf.doCheck()
	if err != nil {
		return nil, err
	}

	ep := &ParamsTransformer{
		WorkerBase: drivers.Worker(id, name),
		params:     conf.Params,
		remove:     conf.Remove,
		errorType:  conf.ErrorType,
	}

	return ep, nil
}
