package anthropic

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
	apikey  string
	version string
	eocontext.BalanceHandler
}

func newConverterDriver(cfg *Config) (convert.IConverterDriver, error) {
	var balanceHandler eocontext.BalanceHandler
	var err error
	if cfg.Base != "" {
		balanceHandler, err = convert.NewBalanceHandler("", cfg.Base, 0)
		if err != nil {
			return nil, err
		}
	}
	return &converterDriver{
		apikey:         cfg.APIKey,
		version:        cfg.Version,
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

func (c *converterDriver) GetConverter(model string) (convert.IConverter, bool) {
	converter, ok := modelConvert[model]
	if !ok {
		return nil, false
	}

	return &Converter{
		balanceHandler: c.BalanceHandler,
		converter:      converter,
		apikey:         c.apikey,
		version:        c.version,
	}, true
}

type Converter struct {
	apikey         string
	version        string
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
	httpContext.Proxy().Header().SetHeader("x-api-key", c.apikey)
	httpContext.Proxy().Header().SetHeader("anthropic-version", c.version)

	return c.converter.RequestConvert(httpContext, extender)
}

func (c *Converter) ResponseConvert(ctx eocontext.EoContext) error {
	return c.converter.ResponseConvert(ctx)
}
