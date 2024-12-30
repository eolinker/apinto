package ai_provider

import (
	"context"

	"github.com/eolinker/apinto/convert"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

var _ eosc.IWorker = (*executor)(nil)

type executor struct {
	drivers.WorkerBase
	provider string
	model    string
	ctx      context.Context
	cancel   context.CancelFunc
	convert.IKeyPool
}

func (e *executor) Start() error {
	return nil
}

func (e *executor) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	cfg, err := checkConfig(conf)
	if err != nil {
		return err
	}
	if err := e.reset(cfg); err != nil {
		return err
	}
	return nil
}

func (e *executor) reset(cfg *Config) error {
	kr, err := newKeyPool(e.ctx, cfg)
	if err != nil {
		return err
	}
	if e.IKeyPool != nil {
		e.IKeyPool.Close()
	}
	e.model = cfg.Model
	e.provider = cfg.Provider
	e.IKeyPool = kr
	return nil
}

func (e *executor) Stop() error {
	e.cancel()
	return nil
}

func (e *executor) CheckSkill(skill string) bool {
	return convert.CheckSkill(skill)
}
