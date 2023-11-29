package gzip

import (
	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/drivers"
)

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

	err := Check(conf, workers)
	if err != nil {
		return nil, err
	}
	c := &Gzip{
		WorkerBase: drivers.Worker(id, name),
		conf:       conf,
	}
	return c, nil
}
