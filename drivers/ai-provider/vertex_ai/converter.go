package vertex_ai

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/eolinker/apinto/convert"
	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
	"golang.org/x/oauth2"
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
	token     *oauth2.Token
	projectId string
	location  string
	jwtData   []byte
}

func newConverterDriver(conf *Config) (convert.IConverterDriver, error) {
	jwtData, err := base64.RawStdEncoding.DecodeString(conf.ServiceAccountKey)
	token, err := newToken(context.Background(), jwtData)
	if err != nil {
		return nil, err
	}
	base := fmt.Sprintf("https://%s-aiplatform.googleapis.com", conf.Location)
	if conf.Base != "" {
		base = conf.Base
	}
	balanceHandler, err := convert.NewBalanceHandler("", base, 0)
	if err != nil {
		return nil, err
	}
	return &converterDriver{
		BalanceHandler: balanceHandler,
		token:          token,
		projectId:      conf.ProjectID,
		location:       conf.Location,
		jwtData:        jwtData,
	}, nil
}

func (e *converterDriver) GetConverter(model string) (convert.IConverter, bool) {
	converter, ok := modelConvert[model]
	if !ok {
		return nil, false
	}

	return &Converter{balanceHandler: e.BalanceHandler, converter: converter, token: e.token, jwtData: e.jwtData, projectId: e.projectId, location: e.location, model: model}, true
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
			if modelCfg.MaxOutputTokens > 0 {
				result["maxOutputTokens"] = modelCfg.MaxOutputTokens
			}
			if modelCfg.Temperature > 0 {
				result["temperature"] = modelCfg.Temperature
			}
			if modelCfg.TopP > 0 {
				result["topP"] = modelCfg.TopP
			}
			if modelCfg.TopK > 1 {
				result["topK"] = modelCfg.TopK
			}
			if modelCfg.FrequencyPenalty > 0 {
				result["frequencyPenalty"] = modelCfg.FrequencyPenalty
			}
			if modelCfg.PresencePenalty > 0 {
				result["presencePenalty"] = modelCfg.PresencePenalty
			}
		}
		return result, nil
	}, true
}

type Converter struct {
	balanceHandler eocontext.BalanceHandler
	model          string
	token          *oauth2.Token
	jwtData        []byte
	projectId      string
	location       string
	converter      convert.IChildConverter
}

func (c *Converter) RequestConvert(ctx eocontext.EoContext, extender map[string]interface{}) error {
	if c.balanceHandler != nil {
		ctx.SetBalance(c.balanceHandler)
	}
	if c.token.Expiry.Before(time.Now()) {
		t, err := newToken(ctx.Context(), c.jwtData)
		if err != nil {
			return err
		}
		c.token = t
	}
	httpContext, err := http_context.Assert(ctx)
	if err != nil {
		return err
	}
	httpContext.Proxy().Header().SetHeader("Authorization", fmt.Sprintf("Bearer %s", c.token.AccessToken))
	httpContext.Proxy().URI().SetPath(fmt.Sprintf(c.converter.Endpoint(), c.projectId, c.location, c.model))
	return c.converter.RequestConvert(httpContext, extender)
}

func (c *Converter) ResponseConvert(ctx eocontext.EoContext) error {
	return c.converter.ResponseConvert(ctx)
}
