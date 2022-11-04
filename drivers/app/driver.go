package app

import (
	"errors"
	"github.com/eolinker/apinto/application"
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

//Create 创建驱动实例
func Create(id, name string, v *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	cfg, err := checkConfig(v)
	if err != nil {
		return nil, err
	}
	_, _, err = createFilters(id, cfg.Auth)
	if err != nil {
		return nil, err
	}
	a := &app{
		WorkerBase: drivers.Worker(id, name),
	}
	err = a.set(cfg)

	return a, err
}

func checkConfig(v interface{}) (*Config, error) {
	conf, ok := v.(*Config)
	if !ok {
		return nil, eosc.ErrorConfigType
	}
	if conf.Anonymous && len(conf.Auth) > 0 {
		return nil, errors.New("it is anonymous app,auths should be empty")
	}
	if conf.Anonymous && len(conf.Auth) > 0 {
		return nil, errors.New("it is anonymous app,auths should be empty")
	}
	for _, a := range conf.Auth {
		err := application.CheckPosition(a.Position)
		if err != nil {
			return nil, err
		}
	}
	for _, a := range conf.Additional {
		err := application.CheckPosition(a.Position)
		if err != nil {
			return nil, err
		}
	}

	return conf, nil
}
