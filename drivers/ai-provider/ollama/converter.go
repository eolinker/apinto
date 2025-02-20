package ollama

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
	apikey string
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
		BalanceHandler: balanceHandler,
	}, nil
}

func (e *converterDriver) GetConverter(model string) (convert.IConverter, bool) {
	return &Converter{balanceHandler: e.BalanceHandler, converter: NewChat()}, true
}

func (e *converterDriver) GetModel(model string) (convert.FGenerateConfig, bool) {
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
			data, err := json.Marshal(modelCfg)
			if err != nil {
				log.Errorf("marshal config error: %v, cfg: %s", err, cfg)
				return result, err
			}
			options := make(map[string]interface{})
			err = json.Unmarshal(data, &options)
			if err != nil {
				return result, err
			}
			result["options"] = options

		}
		return result, nil
	}, true
}

type Converter struct {
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

	return c.converter.RequestConvert(httpContext, extender)
}

func (c *Converter) ResponseConvert(ctx eocontext.EoContext) error {
	return c.converter.ResponseConvert(ctx)
}
