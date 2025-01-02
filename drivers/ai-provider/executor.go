package ai_provider

import (
	"context"

	scope_manager "github.com/eolinker/apinto/scope-manager"

	"github.com/eolinker/apinto/convert"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

var _ eosc.IWorker = (*executor)(nil)
var _ convert.IKeyPool = (*executor)(nil)

type executor struct {
	drivers.WorkerBase
	provider    string
	model       string
	modelConfig map[string]interface{}
	priority    int
	ctx         context.Context
	cancel      context.CancelFunc
	pool        *keyPool
}

func (e *executor) ModelConfig() map[string]interface{} {
	return e.modelConfig
}

func (e *executor) Health() bool {
	if e.pool == nil {
		return false
	}
	return e.pool.Health()
}

func (e *executor) Down() {
	if e.pool == nil {
		return
	}
	e.pool.Down()
}

func (e *executor) Priority() int {
	return e.priority
}

func (e *executor) Provider() string {
	return e.provider
}

func (e *executor) Model() string {
	return e.model
}

func (e *executor) Selector() convert.IKeySelector {
	return e.pool.Selector()
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
	kr, extender, err := newKeyPool(e.ctx, cfg)
	if err != nil {
		return err
	}
	if e.pool != nil {
		e.pool.Close()
	}
	e.priority = cfg.Priority
	e.model = cfg.Model
	e.provider = cfg.Provider
	e.pool = kr
	e.modelConfig = extender
	scope_manager.Set(e.Id(), e, "ai_keys")
	return nil
}

func (e *executor) Stop() error {
	e.cancel()
	scope_manager.Del(e.Id())
	return nil
}

func (e *executor) CheckSkill(skill string) bool {
	return convert.CheckSkill(skill)
}
