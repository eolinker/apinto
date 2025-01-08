package ai_key

import (
	"context"
	"errors"

	"github.com/eolinker/apinto/convert"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

var _ eosc.IWorker = (*executor)(nil)

type executor struct {
	drivers.WorkerBase
	ctx      context.Context
	cancel   context.CancelFunc
	provider string
	key      convert.IKeyResource
}

func (e *executor) Start() error {
	return nil
}

func (e *executor) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	cfg, err := checkConfig(conf)
	if err != nil {
		return err
	}

	return e.reset(cfg)
}

func (e *executor) reset(conf *Config) error {
	factory, has := providerManager.Get(conf.Provider)
	if !has {
		return errors.New("provider not found")
	}

	cv, err := factory.Create(conf.Config)
	if err != nil {
		return err
	}
	k := newKey(e.Id(), e.Name(), conf.Expired, conf.Priority, cv)

	e.key = k
	e.provider = conf.Provider
	convert.SetKeyResource(e.provider, e.key)
	return nil
}

func (e *executor) Stop() error {
	convert.DelKeyResource(e.provider, e.Id())
	return nil
}

func (e *executor) CheckSkill(skill string) bool {
	return convert.CheckKeySourceSkill(skill)
}
