package minimax

import (
	"encoding/json"

	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
)

var _ ai_convert.IConverterFactory = &convertFactory{}

type convertFactory struct {
}

func (c *convertFactory) Create(cfg string) (ai_convert.IConverterDriver, error) {
	var tmp Config
	err := json.Unmarshal([]byte(cfg), &tmp)
	if err != nil {
		return nil, err
	}
	return newConverterDriver(&tmp)
}

var _ ai_convert.IConverterDriver = &converterDriver{}

type converterDriver struct {
	apikey string
}

func newConverterDriver(cfg *Config) (ai_convert.IConverterDriver, error) {
	return &converterDriver{
		apikey: cfg.APIKey,
	}, nil
}

func (e *converterDriver) GetConverter(model string) (ai_convert.IConverter, bool) {
	converter, ok := modelConvert[model]
	if !ok {
		return nil, false
	}

	return &Converter{converter: converter, apikey: e.apikey}, true
}

func (e *converterDriver) GetModel(model string) (ai_convert.FGenerateConfig, bool) {
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
			modelCfg := ai_convert.MapToStruct[ModelConfig](tmp)
			if modelCfg.MaxTokens >= 1 {
				result["max_tokens"] = modelCfg.MaxTokens
			}

			if modelCfg.Temperature > 0 {
				result["temperature"] = modelCfg.Temperature
			}
			if modelCfg.TopP > 0 {
				result["top_p"] = modelCfg.TopP
			}
		}
		return result, nil
	}, true
}

type Converter struct {
	apikey    string
	converter ai_convert.IConverter
}

func (c *Converter) RequestConvert(ctx eocontext.EoContext, extender map[string]interface{}) error {
	httpContext, err := http_context.Assert(ctx)
	if err != nil {
		return err
	}
	httpContext.Proxy().Header().SetHeader("Authorization", "Bearer "+c.apikey)

	return c.converter.RequestConvert(httpContext, extender)
}

func (c *Converter) ResponseConvert(ctx eocontext.EoContext) error {
	return c.converter.ResponseConvert(ctx)
}
