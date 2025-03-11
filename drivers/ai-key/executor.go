package ai_key

import (
	"context"
	"errors"

	ai_convert "github.com/eolinker/apinto/ai-convert"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

var _ eosc.IWorker = (*executor)(nil)

type executor struct {
	drivers.WorkerBase
	ctx      context.Context
	cancel   context.CancelFunc
	provider string
	key      ai_convert.IKeyResource
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
	createFunc, has := ai_convert.GetConverterCreateFunc(conf.Provider)
	if !has {
		createFunc, has = ai_convert.GetConverterCreateFunc("customize-openai")
		if !has {
			return errors.New("provider not found")
		}
	}

	cv, err := createFunc(conf.Config)
	if err != nil {
		return err
	}
	k := newKey(e.Id(), e.Name(), conf.Expired, conf.Priority, cv)

	e.key = k
	e.provider = conf.Provider
	ai_convert.SetKeyResource(e.provider, e.key)
	return nil
}

func (e *executor) Stop() error {
	ai_convert.DelKeyResource(e.provider, e.Id())
	return nil
}

func (e *executor) CheckSkill(skill string) bool {
	return ai_convert.CheckKeySourceSkill(skill)
}
