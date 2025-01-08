package baichuan

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
	apikey string
	eocontext.BalanceHandler
}

func newConverterDriver(cfg *Config) (convert.IConverterDriver, error) {
	var balanceHandler eocontext.BalanceHandler
	return &converterDriver{
		apikey:         cfg.APIKey,
		BalanceHandler: balanceHandler,
	}, nil

}

func (c *converterDriver) GetModel(model string) (convert.FGenerateConfig, bool) {
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
			if modelCfg.MaxTokens >= 1 {
				result["max_tokens"] = modelCfg.MaxTokens
			}

			result["temperature"] = modelCfg.Temperature
			result["top_p"] = modelCfg.TopP
		}
		return result, nil
	}, true
}

func (c *converterDriver) GetConverter(model string) (convert.IConverter, bool) {
	converter, ok := modelConvert[model]
	if !ok {
		return nil, false
	}

	return &Converter{balanceHandler: c.BalanceHandler, converter: converter, apikey: c.apikey}, true
}

type Converter struct {
	apikey         string
	balanceHandler eocontext.BalanceHandler
	converter      convert.IConverter
}

func (c *Converter) RequestConvert(ctx eocontext.EoContext, extender map[string]interface{}) error {
	if c.balanceHandler != nil {
		ctx.SetBalance(c.balanceHandler)
	}
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
