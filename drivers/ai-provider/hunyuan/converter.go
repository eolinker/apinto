package hunyuan

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
	eocontext.BalanceHandler

	secretId string
	secret   string
}

func newConverterDriver(cfg *Config) (convert.IConverterDriver, error) {
	return &converterDriver{
		secret:   cfg.SecretKey,
		secretId: cfg.SecretID,
	}, nil
}

func (e *converterDriver) GetConverter(model string) (convert.IConverter, bool) {
	converter, ok := modelConvert[model]
	if !ok {
		return nil, false
	}

	return &Converter{converter: converter, secretID: e.secretId, secretKey: e.secret}, true
}

func (e *converterDriver) GetModel(model string) (convert.FGenerateConfig, bool) {
	if _, ok := modelConvert[model]; !ok {
		return nil, false
	}
	return func(cfg string) (map[string]interface{}, error) {

		result := map[string]interface{}{
			"Model": model,
		}
		if cfg != "" {
			tmp := make(map[string]interface{})
			if err := json.Unmarshal([]byte(cfg), &tmp); err != nil {
				log.Errorf("unmarshal config error: %v, cfg: %s", err, cfg)
				return result, nil
			}
			modelCfg := convert.MapToStruct[ModelConfig](tmp)

			result["EnableEnhancement"] = modelCfg.EnableEnhance

			result["Temperature"] = modelCfg.Temperature
			result["TopP"] = modelCfg.TopP
		}
		return result, nil
	}, true
}

type Converter struct {
	secretID  string
	secretKey string
	converter convert.IConverter
}

func (c *Converter) RequestConvert(ctx eocontext.EoContext, extender map[string]interface{}) error {

	httpContext, err := http_context.Assert(ctx)
	if err != nil {
		return err
	}

	err = c.converter.RequestConvert(httpContext, extender)
	if err != nil {
		return err
	}
	return Sign(httpContext, c.secretID, c.secretKey)
}

func (c *Converter) ResponseConvert(ctx eocontext.EoContext) error {
	return c.converter.ResponseConvert(ctx)
}
