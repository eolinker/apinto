package ai_model

import (
	"encoding/json"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

var _ eosc.IWorker = (*executor)(nil)

type executor struct {
	drivers.WorkerBase
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
	tmp := make(map[string]string)
	err := json.Unmarshal([]byte(conf.AccessConfig), &tmp)
	if err != nil {
		return err
	}
	accessConfigManager.Set(e.Name(), &modelAccessConfig{
		provider:     conf.Provider,
		model:        conf.Model,
		accessConfig: tmp,
	})

	return nil
}

func (e *executor) Stop() error {
	accessConfigManager.Del(e.Name())
	return nil
}

func (e *executor) CheckSkill(skill string) bool {
	return false
}

type modelAccessConfig struct {
	provider     string
	model        string
	accessConfig map[string]string
}

func (m *modelAccessConfig) Provider() string {
	return m.provider
}

func (m *modelAccessConfig) Model() string {
	return m.model
}

func (m *modelAccessConfig) Config() map[string]string {
	return m.accessConfig
}
