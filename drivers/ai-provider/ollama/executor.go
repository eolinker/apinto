package ollama

import (
	"fmt"

	"github.com/eolinker/apinto/drivers"

	"github.com/eolinker/eosc"
)

var (
	_ ai_convert.IConverterDriver = (*executor)(nil)
)

type executor struct {
	drivers.WorkerBase
	ai_convert.IConverterDriver
}

func (e *executor) Start() error {
	return nil
}

func (e *executor) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	cfg, ok := conf.(*Config)
	if !ok {
		return fmt.Errorf("invalid config")
	}

	return e.reset(cfg, workers)
}

func (e *executor) reset(conf *Config, workers map[eosc.RequireId]eosc.IWorker) error {
	d, err := newConverterDriver(conf)
	if err != nil {
		return err
	}

	e.IConverterDriver = d

	return nil
}

func (e *executor) Stop() error {
	e.IConverterDriver = nil
	return nil
}

func (e *executor) CheckSkill(skill string) bool {
	return ai_convert.CheckKeySourceSkill(skill)
}

type ModelConfig struct {
	Mirostat      int     `json:"mirostat,omitempty"`
	MirostatEta   float64 `json:"mirostat_eta,omitempty"`
	MirostatTau   float64 `json:"mirostat_tau,omitempty"`
	NumCtx        int     `json:"num_ctx,omitempty"`
	RepeatLastN   int     `json:"repeat_last_n,omitempty"`
	RepeatPenalty float64 `json:"repeat_penalty,omitempty"`
	Seed          int     `json:"seed,omitempty"`
	NumPredict    int     `json:"num_predict,omitempty"`
	TopK          int     `json:"top_k,omitempty"`
	TopP          float64 `json:"top_p,omitempty"`
	MinP          float64 `json:"min_p,omitempty"`
}
