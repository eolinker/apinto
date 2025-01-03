package ai_provider

import (
	"fmt"

	eoscContext "github.com/eolinker/eosc/eocontext"

	"github.com/eolinker/apinto/convert"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

var _ eosc.IWorker = (*executor)(nil)

var _ convert.IProvider = (*executor)(nil)

type executor struct {
	drivers.WorkerBase
	provider       string
	model          string
	modelConfig    map[string]interface{}
	priority       int
	disable        bool
	balanceHandler eoscContext.BalanceHandler
}

func (e *executor) BalanceHandler() eoscContext.BalanceHandler {
	return e.balanceHandler
}

func (e *executor) Health() bool {
	return !e.disable
}

func (e *executor) Down() {
	e.disable = true
}

func (e *executor) ModelConfig() map[string]interface{} {
	return e.modelConfig
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
	factory, has := providerManager.Get(cfg.Provider)
	if !has {
		return fmt.Errorf("provider not found")
	}
	cv, err := factory.Create("{}")
	if err != nil {
		return err
	}
	fn, has := cv.GetModel(cfg.Model)
	if !has {
		return fmt.Errorf("default model not found")
	}
	extender, err := fn(cfg.ModelConfig)
	if err != nil {
		return err
	}
	if cfg.Base != "" {
		balanceHandler, err := convert.NewBalanceHandler("", cfg.Base, 0)
		if err != nil {
			return err
		}
		e.balanceHandler = balanceHandler
	}
	e.priority = cfg.Priority
	e.model = cfg.Model
	e.provider = cfg.Provider
	e.modelConfig = extender
	return nil
}

func (e *executor) Stop() error {
	convert.DelProvider(e.provider)
	return nil
}

func (e *executor) CheckSkill(skill string) bool {
	return convert.CheckKeySourceSkill(skill)
}
