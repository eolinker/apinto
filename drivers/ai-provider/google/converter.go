package google

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
	eocontext.BalanceHandler
	apikey string
}

func newConverterDriver(cfg *Config) (ai_convert.IConverterDriver, error) {
	var balanceHandler eocontext.BalanceHandler
	var err error
	if cfg.Base != "" {
		balanceHandler, err = ai_convert.NewBalanceHandler("", cfg.Base, 0)
		if err != nil {
			return nil, err
		}
	}
	return &converterDriver{
		apikey:         cfg.APIKey,
		BalanceHandler: balanceHandler,
	}, nil
}

func (e *converterDriver) GetConverter(model string) (ai_convert.IConverter, bool) {
	converter, ok := modelConvert[model]
	if !ok {
		return nil, false
	}

	return &Converter{balanceHandler: e.BalanceHandler, converter: converter, apikey: e.apikey}, true
}

func (e *converterDriver) GetModel(model string) (ai_convert.FGenerateConfig, bool) {
	if _, ok := modelConvert[model]; !ok {
		return nil, false
	}
	return func(cfg string) (map[string]interface{}, error) {
		result := map[string]interface{}{}
		if cfg != "" {
			tmp := make(map[string]interface{})
			if err := json.Unmarshal([]byte(cfg), &tmp); err != nil {
				log.Errorf("unmarshal config error: %v, cfg: %s", err, cfg)
				return result, nil
			}
			modelCfg := ai_convert.MapToStruct[ModelConfig](tmp)
			generationConfig := make(map[string]interface{})
			generationConfig["maxOutputTokens"] = modelCfg.MaxOutputTokens
			generationConfig["temperature"] = modelCfg.Temperature
			generationConfig["topP"] = modelCfg.TopP
			generationConfig["topK"] = modelCfg.TopK
			result["generationConfig"] = generationConfig
		}
		return result, nil
	}, true
}

type Converter struct {
	apikey         string
	balanceHandler eocontext.BalanceHandler
	converter      ai_convert.IConverter
}

func (c *Converter) RequestConvert(ctx eocontext.EoContext, extender map[string]interface{}) error {
	if c.balanceHandler != nil {
		ctx.SetBalance(c.balanceHandler)
	}
	httpContext, err := http_context.Assert(ctx)
	if err != nil {
		return err
	}
	httpContext.Proxy().URI().SetQuery("key", c.apikey)

	return c.converter.RequestConvert(httpContext, extender)
}

func (c *Converter) ResponseConvert(ctx eocontext.EoContext) error {
	return c.converter.ResponseConvert(ctx)
}
