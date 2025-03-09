package bedrock

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	ai_convert "github.com/eolinker/apinto/ai-convert"

	"github.com/aws/aws-sdk-go/aws/credentials"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"

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

type basicConfig struct {
	signer *v4.Signer
	region string
	eocontext.BalanceHandler
}

type converterDriver struct {
	cfg *basicConfig
	eocontext.BalanceHandler
}

func newConverterDriver(cfg *Config) (ai_convert.IConverterDriver, error) {
	base := fmt.Sprintf("https://bedrock-runtime.%s.amazonaws.com", cfg.Region)
	balanceHandler, err := ai_convert.NewBalanceHandler("", base, 0)
	if err != nil {
		return nil, err
	}
	return &converterDriver{
		cfg: &basicConfig{
			signer:         v4.NewSigner(credentials.NewStaticCredentials(cfg.AccessKey, cfg.SecretKey, "")),
			region:         cfg.Region,
			BalanceHandler: balanceHandler,
		},
		BalanceHandler: balanceHandler,
	}, nil

}

func (c *converterDriver) GetModel(model string) (ai_convert.FGenerateConfig, bool) {
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
			if modelCfg.MaxTokens >= 1 {
				result["maxTokens"] = modelCfg.MaxTokens
			}
			result["temperature"] = modelCfg.Temperature
			result["topP"] = modelCfg.TopP
		}
		return result, nil
	}, true
}

func (c *converterDriver) GetConverter(model string) (ai_convert.IConverter, bool) {
	converter, ok := modelConvert[model]
	if !ok {
		return nil, false
	}

	return &Converter{
		converter:   converter,
		model:       model,
		basicConfig: c.cfg,
	}, true
}

type Converter struct {
	converter ai_convert.IConverter
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
