package bedrock

import (
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"

	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"

	"github.com/eolinker/eosc/log"

	"github.com/eolinker/apinto/drivers"

	http_context "github.com/eolinker/eosc/eocontext/http-context"

	ai_provider "github.com/eolinker/apinto/drivers/ai-provider"

	"github.com/eolinker/apinto/convert"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
)

var (
	//go:embed bedrock.yaml
	providerContent []byte
	//go:embed *
	providerDir  embed.FS
	modelConvert = make(map[string]convert.IConverter)

	_ convert.IConverterDriver = (*executor)(nil)
)

func init() {
	models, err := ai_provider.LoadModels(providerContent, providerDir)
	if err != nil {
		panic(err)
	}
	for key, value := range models {
		if value.ModelProperties != nil {
			if v, ok := modelModes[value.ModelProperties.Mode]; ok {
				modelConvert[key] = v(value.Model)
			}
		}
	}
}

type Converter struct {
	converter convert.IConverter
	model     string
	*basicConfig
}

func (c *Converter) RequestConvert(ctx eocontext.EoContext, extender map[string]interface{}) error {
	if c.BalanceHandler != nil {
		ctx.SetBalance(c.BalanceHandler)
	}
	httpContext, err := http_context.Assert(ctx)
	if err != nil {
		return err
	}

	err = c.converter.RequestConvert(httpContext, extender)
	if err != nil {
		return err
	}
	body, _ := httpContext.Proxy().Body().RawBody()
	headers, err := signRequest(c.signer, c.region, c.model, http.Header{}, string(body))
	if err != nil {
		return err
	}
	for k, v := range headers {

		httpContext.Proxy().Header().SetHeader(k, strings.Join(v, ";"))
	}
	//httpContext.Proxy().Header().SetHeader("Authorization", authorization)
	//httpContext.Proxy().Header().SetHeader("X-Amz-Date", date)
	return nil
}

func (c *Converter) ResponseConvert(ctx eocontext.EoContext) error {
	return c.converter.ResponseConvert(ctx)
}

type executor struct {
	drivers.WorkerBase
	cfg *basicConfig
}

type basicConfig struct {
	signer *v4.Signer
	region string
	eocontext.BalanceHandler
}

func (e *executor) GetConverter(model string) (convert.IConverter, bool) {
	converter, ok := modelConvert[model]
	if !ok {
		return nil, false
	}

	return &Converter{
		converter:   converter,
		model:       model,
		basicConfig: e.cfg,
	}, true
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
			modelCfg := ai_provider.MapToStruct[ModelConfig](tmp)
			if modelCfg.MaxTokens >= 1 {
				result["maxTokens"] = modelCfg.MaxTokens
			}
			result["temperature"] = modelCfg.Temperature
			result["topP"] = modelCfg.TopP
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
	base := fmt.Sprintf("https://bedrock-runtime.%s.amazonaws.com", conf.Region)
	u, err := url.Parse(base)
	if err != nil {
		return err
	}
	hosts := strings.Split(u.Host, ":")
	ip := hosts[0]
	port := 80
	if u.Scheme == "https" {
		port = 443
	}
	if len(hosts) > 1 {
		port, _ = strconv.Atoi(hosts[1])
	}
	e.cfg = &basicConfig{
		signer:         v4.NewSigner(credentials.NewStaticCredentials(conf.AccessKey, conf.SecretKey, "")),
		region:         conf.Region,
		BalanceHandler: ai_provider.NewBalanceHandler(u.Scheme, 0, []eocontext.INode{ai_provider.NewBaseNode(e.Id(), ip, port)}),
	}
	convert.Set(e.Id(), e)

	return nil
}

func (e *executor) Stop() error {
	e.cfg = nil
	convert.Del(e.Id())
	return nil
}

func (e *executor) CheckSkill(skill string) bool {
	return convert.CheckSkill(skill)
}

type ModelConfig struct {
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
	TopP        float64 `json:"top_p"`
}

func signRequest(signer *v4.Signer, region string, model string, headers http.Header, body string) (http.Header, error) {
	request, err := http.NewRequest(http.MethodPost, fmt.Sprintf("https://bedrock-runtime.%s.amazonaws.com/model/%s/converse", region, model), nil)
	if err != nil {
		return nil, err
	}
	request.Header = headers.Clone()

	_, err = signer.Sign(request, strings.NewReader(body), "bedrock", region, time.Now())
	if err != nil {
		return nil, err
	}
	return request.Header, nil

}
