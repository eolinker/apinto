package bailianyun

import (
	"fmt"
	ai_convert "github.com/eolinker/apinto/ai-convert"
	context_label "github.com/eolinker/apinto/common/context-label"
	eoscContext "github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"net/url"
	"strings"
	"time"
)

func init() {
	driverCreate.Set(ai_convert.ModelTypeImageGeneration, func(mt ai_convert.ModelType, c *Config) (ai_convert.IConverterDriver, error) {
		return NewImageGeneration(provider, c.APIKey, c.BaseUrl, mt, time.Minute*10)
	})
}

var _ ai_convert.IConverterDriver = (*ImageGeneration)(nil)

type ImageRequest struct {
	Model string `json:"model"`
}

type ImageResponse struct {
}

const (
	imagePath = "/services/aigc/multimodal-generation/generation"
)

func NewImageGeneration(provider string, apikey string, baseUrl string, modelType ai_convert.ModelType, timeout time.Duration) (ai_convert.IConverterDriver, error) {
	c := &ImageGeneration{
		provider:  provider,
		apikey:    apikey,
		modelType: modelType,
		path:      imagePath,
	}

	balanceHandler, err := ai_convert.NewBalanceHandler(apikey, baseUrl, timeout)
	if err != nil {
		return nil, err
	}
	c.balanceHandler = balanceHandler
	u, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}

	if strings.TrimSuffix(u.Path, "/") == "" {
		c.path = "/api/v1" + c.path
	} else {
		c.path = fmt.Sprintf("%s%s", strings.TrimSuffix(u.Path, "/"), c.path)
	}
	return c, nil
}

type ImageGeneration struct {
	provider       string
	apikey         string
	modelType      ai_convert.ModelType
	path           string
	balanceHandler eoscContext.BalanceHandler
}

func (i *ImageGeneration) Provider() string {
	return i.provider
}

func (i *ImageGeneration) ModelType() ai_convert.ModelType {
	return i.modelType
}

func (i *ImageGeneration) RequestConvert(ctx eoscContext.EoContext, extender map[string]interface{}) error {
	httpContext, err := http_service.Assert(ctx)
	if err != nil {
		return err
	}

	if i.apikey != "" {
		httpContext.Proxy().Header().SetHeader("Authorization", "Bearer "+i.apikey)
	}
	httpContext.Proxy().URI().SetPath(i.path)
	if i.balanceHandler != nil {
		ctx.SetBalance(i.balanceHandler)
	}
	context_label.SetDisableStream(ctx, true)

	return nil
}

func (i *ImageGeneration) ResponseConvert(ctx eoscContext.EoContext) error {
	return nil
}
