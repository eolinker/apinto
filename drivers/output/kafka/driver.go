package kafka

import (
	"reflect"

	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/drivers"
)

type Driver struct {
	configType reflect.Type
}

func Check(v *Config, workers map[eosc.RequireId]eosc.IWorker) error {
	_, err := v.doCheck()

	return err
}

func check(v interface{}) (*ProducerConfig, error) {
	conf, err := drivers.Assert[Config](v)
	if err != nil {
		return nil, err
	}

	pConf, err := conf.doCheck()
	if err != nil {
		return nil, err
	}
	return pConf, nil

}

func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	cfg, err := conf.doCheck()
	if err != nil {
		return nil, err
	}

	worker := &Output{
		WorkerBase: drivers.Worker(id, name),
		producer:   nil,
		config:     cfg,
	}

	return worker, err
}
