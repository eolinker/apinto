package wenxin

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
	apikey    string
	secretKey string
}

func newConverterDriver(cfg *Config) (convert.IConverterDriver, error) {
	return &converterDriver{
		apikey:    cfg.APIKey,
		secretKey: cfg.SecretKey,
	}, nil
}

func (e *converterDriver) GetConverter(model string) (convert.IConverter, bool) {
	converter, ok := modelConvert[model]
	if !ok {
		return nil, false
	}

	return &Converter{converter: converter, apikey: e.apikey, secret: e.secretKey}, true
}

func (e *converterDriver) GetModel(model string) (convert.FGenerateConfig, bool) {
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
			if modelCfg.MaxTokens >= 2 {
				result["max_output_tokens"] = modelCfg.MaxTokens
			}
			result["disable_search"] = modelCfg.DisableSearch
			if modelCfg.PresencePenalty >= 1 && modelCfg.PresencePenalty <= 2 {
				result["penalty_score"] = modelCfg.PresencePenalty
			}
			if modelCfg.Temperature > 0 && modelCfg.Temperature <= 1 {
				result["temperature"] = modelCfg.Temperature
			}
			if modelCfg.TopP > 0 && modelCfg.TopP <= 1 {
				result["top_p"] = modelCfg.TopP
			}

			if modelCfg.ResponseFormat == "" {
				modelCfg.ResponseFormat = "text"
			}
			result["response_format"] = modelCfg.ResponseFormat
		}
		return result, nil
	}, true
}

type Converter struct {
	apikey    string
	secret    string
	converter convert.IConverter
}

func (c *Converter) RequestConvert(ctx eocontext.EoContext, extender map[string]interface{}) error {

	httpContext, err := http_context.Assert(ctx)
	if err != nil {
		return err
	}
	//httpContext.Proxy().Header().SetHeader("Authorization", "Bearer "+c.apikey)
	//err = Sign(c.apikey, c.secret, httpContext)
	//if err != nil {
	//	return err
	//}
	token, err := getToken(c.apikey, c.secret)
	if err != nil {
		return err
	}
	httpContext.Proxy().URI().SetQuery("access_token", token.AccessToken)

	return c.converter.RequestConvert(httpContext, extender)
}

func (c *Converter) ResponseConvert(ctx eocontext.EoContext) error {
	return c.converter.ResponseConvert(ctx)
}
