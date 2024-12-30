package vertex_ai

import (
	"context"
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"golang.org/x/oauth2/google"

	"golang.org/x/oauth2"

	"github.com/eolinker/eosc/log"

	"github.com/eolinker/apinto/drivers"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
	dns "google.golang.org/api/dns/v1beta2"

	"github.com/eolinker/apinto/convert"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
)

var (
	//go:embed vertex_ai.yaml
	providerContent []byte
	//go:embed *
	providerDir  embed.FS
	modelConvert = make(map[string]convert.IChildConverter)

	_      convert.IConverterDriver = (*executor)(nil)
	scopes                          = []string{
		dns.CloudPlatformReadOnlyScope,
		dns.CloudPlatformScope,
	}
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

type executor struct {
	drivers.WorkerBase
	eocontext.BalanceHandler
	token     *oauth2.Token
	projectId string
	location  string
	jwtData   []byte
}

func (e *executor) GetConverter(model string) (convert.IConverter, bool) {
	converter, ok := modelConvert[model]
	if !ok {
		return nil, false
	}

	return &Converter{balanceHandler: e.BalanceHandler, converter: converter, token: e.token, jwtData: e.jwtData, projectId: e.projectId, location: e.location, model: model}, true
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
	jwtData, err := base64.RawStdEncoding.DecodeString(conf.ServiceAccountKey)
	token, err := newToken(context.Background(), jwtData)
	if err != nil {
		return err
	}
	base := fmt.Sprintf("https://%s-aiplatform.googleapis.com", conf.Location)
	if conf.Base != "" {
		base = conf.Base
	}
	balanceHandler, err := convert.NewBalanceHandler(e.Id(), base, 0)
	if err != nil {
		return err
	}
	e.BalanceHandler = balanceHandler
	e.projectId = conf.ProjectID
	e.location = conf.Location
	e.token = token
	e.jwtData = jwtData
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
	MaxOutputTokens  int     `json:"max_tokens"`
	Temperature      float64 `json:"temperature"`
	FrequencyPenalty float64 `json:"frequency_penalty"`
	PresencePenalty  float64 `json:"presence_penalty"`
	TopP             float64 `json:"top_p"`
	TopK             int     `json:"top_k"`
}

func newToken(ctx context.Context, data []byte) (*oauth2.Token, error) {
	cfg, err := google.JWTConfigFromJSON(data, scopes...)
	if err != nil {
		return nil, err
	}
	return cfg.TokenSource(ctx).Token()
}
