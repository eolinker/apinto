package spark

import (
	"encoding/json"

	"github.com/eolinker/apinto/convert"
	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
)

var _ convert.IConverterFactory = &convertFactory{}

type convertFactory struct {
}

func (c *convertFactory) Create(cfg string) (convert.IConverterDriver, error) {
	var tmp Config
	err := json.Unmarshal([]byte(cfg), &tmp)
	if err != nil {
		return nil, err
	}
	return newConverterDriver(&tmp)
}

var _ convert.IConverterDriver = &converterDriver{}

type converterDriver struct {
	apiPassword string
}

func newConverterDriver(cfg *Config) (convert.IConverterDriver, error) {
	return &converterDriver{
		apiPassword: cfg.APIPassword,
	}, nil
}

func (e *converterDriver) GetConverter(model string) (convert.IConverter, bool) {
	converter, ok := modelConvert[model]
	if !ok {
		return nil, false
	}

	return &Converter{converter: converter, apiPassword: e.apiPassword}, true
}

func (e *converterDriver) GetModel(model string) (convert.FGenerateConfig, bool) {
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
			modelCfg := convert.MapToStruct[ModelConfig](tmp)
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
