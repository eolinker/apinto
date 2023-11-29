package syslog

import (
	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/drivers"
)

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

func Create(id, name string, cfg *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	err := cfg.doCheck()

	if err != nil {
		return nil, err
	}

	return &Output{
		WorkerBase: drivers.Worker(id, name),
		config:     cfg,
		writer:     nil,
	}, nil

}
