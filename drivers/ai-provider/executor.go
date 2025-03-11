package ai_provider

import (
	ai_convert "github.com/eolinker/apinto/ai-convert"

	eoscContext "github.com/eolinker/eosc/eocontext"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

var _ eosc.IWorker = (*executor)(nil)

var _ ai_convert.IProvider = (*executor)(nil)

type executor struct {
	drivers.WorkerBase
	provider       string
	model          string
	modelConfig    map[string]interface{}
	priority       int
	disable        bool
	balanceHandler eoscContext.BalanceHandler
}

func (e *executor) GenExtender(cfg string) (map[string]interface{}, error) {
	return ai_convert.TransformData(cfg, providerMapValue)
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
	extender, err := ai_convert.TransformData(cfg.ModelConfig, providerMapValue)
	if err != nil {
		return err
	}
	if cfg.Base != "" {
		balanceHandler, err := ai_convert.NewBalanceHandler("", cfg.Base, 0)
		if err != nil {
			return err
		}
		e.balanceHandler = balanceHandler
	}

	e.priority = cfg.Priority
	e.model = cfg.Model
	e.provider = cfg.Provider
	e.modelConfig = extender
	e.disable = false
	ai_convert.SetProvider(e.Id(), e)
	return nil
}

func (e *executor) Stop() error {
	ai_convert.DelProvider(e.Id())
	return nil
}

func (e *executor) CheckSkill(skill string) bool {
	return ai_convert.CheckKeySourceSkill(skill)
}
