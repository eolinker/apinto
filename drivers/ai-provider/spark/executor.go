package spark

import (
	"embed"
	"encoding/json"
	"fmt"

	"github.com/eolinker/eosc/log"

	"github.com/eolinker/apinto/drivers"

	http_context "github.com/eolinker/eosc/eocontext/http-context"

	ai_provider "github.com/eolinker/apinto/drivers/ai-provider"

	"github.com/eolinker/apinto/convert"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
)

var (
	//go:embed spark.yaml
	providerContent []byte
	//go:embed *
	providerDir  embed.FS
	modelConvert = make(map[string]convert.IConverter)

	_ convert.IConverterDriver = (*executor)(nil)
)

func init() {
	models, err := ai_provider.LoadModels(providerContent, providerDir)
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

type Converter struct {
	apiPassword string
	converter   convert.IConverter
}

func (c *Converter) RequestConvert(ctx eocontext.EoContext, extender map[string]interface{}) error {
	httpContext, err := http_context.Assert(ctx)
	if err != nil {
		return err
	}
	httpContext.Proxy().Header().SetHeader("Authorization", "Bearer "+c.apiPassword)

	return c.converter.RequestConvert(httpContext, extender)
}

func (c *Converter) ResponseConvert(ctx eocontext.EoContext) error {
	return c.converter.ResponseConvert(ctx)
}

type executor struct {
	drivers.WorkerBase
	apiPassword string
}

func (e *executor) GetConverter(model string) (convert.IConverter, bool) {
	converter, ok := modelConvert[model]
	if !ok {
		return nil, false
	}

	return &Converter{converter: converter, apiPassword: e.apiPassword}, true
}

func (e *executor) GetModel(model string) (convert.FGenerateConfig, bool) {
	if _, ok := modelConvert[model]; !ok {
		return nil, false
	}
	return func(cfg string) (map[string]interface{}, error) {

		result := map[string]interface{}{
			"model": model,
		}
		if cfg != "" {
			tmp := make(map[string]interface{})
			if err := json.Unmarshal([]byte(cfg), &tmp); err != nil {
				log.Errorf("unmarshal config error: %v, cfg: %s", err, cfg)
				return result, nil
			}
			modelCfg := ai_provider.MapToStruct[ModelConfig](tmp)
			result["frequency_penalty"] = modelCfg.FrequencyPenalty
			if modelCfg.MaxTokens >= 1 {
				result["max_tokens"] = modelCfg.MaxTokens
			}

			result["presence_penalty"] = modelCfg.PresencePenalty
			result["temperature"] = modelCfg.Temperature
			if modelCfg.TopP > 0 && modelCfg.TopP <= 1 {
				result["top_p"] = modelCfg.TopP
			}
			if modelCfg.TopK >= 1 && modelCfg.TopK <= 6 {
				result["top_k"] = modelCfg.TopK
			}
			if modelCfg.ResponseFormat == "" {
				modelCfg.ResponseFormat = "text"
			}
			result["response_format"] = map[string]interface{}{
				"type": modelCfg.ResponseFormat,
			}
		}
		return result, nil
	}, true
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

	e.apiPassword = conf.APIPassword
	convert.Set(e.Id(), e)

	return nil
}

func (e *executor) Stop() error {
	convert.Del(e.Id())
	return nil
}

func (e *executor) CheckSkill(skill string) bool {
	return convert.CheckSkill(skill)
}

type ModelConfig struct {
	FrequencyPenalty float64 `json:"frequency_penalty"`
	MaxTokens        int     `json:"max_tokens"`
	PresencePenalty  float64 `json:"presence_penalty"`
	ResponseFormat   string  `json:"response_format"`
	Temperature      float64 `json:"temperature"`
	TopP             float64 `json:"top_p"`
	TopK             float64 `json:"top_k"`
}
