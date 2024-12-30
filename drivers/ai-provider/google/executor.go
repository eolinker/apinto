package google

import (
	"embed"
	"encoding/json"
	"fmt"

	"github.com/eolinker/eosc/log"

	"github.com/eolinker/apinto/drivers"

	http_context "github.com/eolinker/eosc/eocontext/http-context"

	"github.com/eolinker/apinto/convert"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
)

var (
	//go:embed google.yaml
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
			if f, ok := modelModes[value.ModelProperties.Mode]; ok {
				modelConvert[key] = f(value.Model)
			}
		}
	}
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
	httpContext.Proxy().URI().SetQuery("key", c.apikey)

	return c.converter.RequestConvert(httpContext, extender)
}

func (c *Converter) ResponseConvert(ctx eocontext.EoContext) error {
	return c.converter.ResponseConvert(ctx)
}

type executor struct {
	drivers.WorkerBase
	eocontext.BalanceHandler
	apikey string
}

func (e *executor) GetConverter(model string) (convert.IConverter, bool) {
	converter, ok := modelConvert[model]
	if !ok {
		return nil, false
	}

	return &Converter{balanceHandler: e.BalanceHandler, converter: converter, apikey: e.apikey}, true
}

func (e *executor) GetModel(model string) (convert.FGenerateConfig, bool) {
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
			modelCfg := convert.MapToStruct[ModelConfig](tmp)
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
	if conf.Base != "" {
		balanceHandler, err := convert.NewBalanceHandler(e.Id(), conf.Base, 0)
		if err != nil {
			return err
		}
		e.BalanceHandler = balanceHandler
	} else {
		e.BalanceHandler = nil
	}
	e.apikey = conf.APIKey
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
	ResponseMimeType string  `json:"response_format"`
	MaxOutputTokens  int     `json:"max_tokens_to_sample"`
	Temperature      float64 `json:"temperature"`
	TopP             float64 `json:"top_p"`
	TopK             int     `json:"top_k"`
}
