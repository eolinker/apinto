package novita

import (
	"embed"
	"fmt"

	"github.com/eolinker/apinto/convert"
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

var (
	//go:embed novita.yaml
	providerContent []byte
	//go:embed *
	providerDir  embed.FS
	modelConvert = make(map[string]convert.IConverter)

	_ convert.IConverterDriver = (*executor)(nil)
)

func init() {
	models, err := convert.LoadModels(providerContent, providerDir)
	if err != nil {
		panic(err)
	}
	for key, value := range models {
		if value.ModelProperties != nil {
			if v, ok := modelModes[value.ModelProperties.Mode]; ok {
				modelConvert[key] = v
			}
		}
	}
}

type executor struct {
	drivers.WorkerBase
	convert.IConverterDriver
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
	if e.IConverterDriver == nil {
		driver, err := newConverterDriver(conf)
		if err != nil {
			return err
		}
		e.IConverterDriver = driver
	}

	return nil
}

func (e *executor) Stop() error {
	e.IConverterDriver = nil
	return nil
}

func (e *executor) CheckSkill(skill string) bool {
	return convert.CheckKeySourceSkill(skill)
}

type ModelConfig struct {
	FrequencyPenalty float64 `json:"frequency_penalty"`
	MaxTokens        int     `json:"max_tokens"`
	PresencePenalty  float64 `json:"presence_penalty"`
	ResponseFormat   string  `json:"response_format"`
	Temperature      float64 `json:"temperature"`
	TopP             float64 `json:"top_p"`
}
