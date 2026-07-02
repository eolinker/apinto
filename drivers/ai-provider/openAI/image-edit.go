package openAI

import (
	"encoding/json"
	"fmt"
	ai_convert "github.com/eolinker/apinto/ai-convert"
	"github.com/eolinker/apinto/common/context-label"
	"github.com/eolinker/eosc"
	eoscContext "github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"net/url"
	"strings"
	"time"
)

func init() {
	driverCreate.Set(ai_convert.ModelTypeImageEdit, func(mt ai_convert.ModelType, c *Config) (ai_convert.IConverterDriver, error) {
		return NewImageEdit(provider, c.APIKey, c.BaseUrl, mt, 10*time.Minute)
	})
}

var _ ai_convert.IConverterDriver = (*ImageGeneration)(nil)

const (
	imageEditPath = "/images/edits"
)

func NewImageEdit(provider string, apikey string, base string, modelType ai_convert.ModelType, timeout time.Duration) (ai_convert.IConverterDriver, error) {
	c := &ImageEdit{
		provider:  provider,
		apikey:    apikey,
		modelType: modelType,
		path:      imageEditPath,
	}
	if base == "" {
		base = baseUrl
	}
	balanceHandler, err := ai_convert.NewBalanceHandler(apikey, base, timeout)
	if err != nil {
		return nil, err
	}
	c.balanceHandler = balanceHandler
	u, err := url.Parse(base)
	if err != nil {
		return nil, err
	}
	if u.Path != "" {
		c.path = fmt.Sprintf("/%s/%s", strings.Trim(u.Path, "/"), strings.Trim(imageGeneratePath, "/"))
	}

	return c, nil
}

type ImageEdit struct {
	provider       string
	apikey         string
	modelType      ai_convert.ModelType
	path           string
	balanceHandler eoscContext.BalanceHandler
}

func (i *ImageEdit) Provider() string {
	return i.provider
}

func (i *ImageEdit) ModelType() ai_convert.ModelType {
	return i.modelType
}

func (i *ImageEdit) RequestConvert(ctx eoscContext.EoContext, extender map[string]interface{}) error {
	httpContext, err := http_service.Assert(ctx)
	if err != nil {
		return err
	}
	body, err := httpContext.Proxy().Body().RawBody()
	if err != nil {
		return err
	}
	chatRequest := eosc.NewBase[ImageRequest](extender)
	err = json.Unmarshal(body, chatRequest)
	if err != nil {
		return fmt.Errorf("unmarshal body error: %v, body: %s", err, string(body))
	}
	if chatRequest.Config.Model == "" {
		chatRequest.Config.Model = ai_convert.GetAIModel(ctx)
	}
	context_label.SetDisableStream(ctx, true)
	httpContext.Proxy().Header().SetHeader("Authorization", "Bearer "+i.apikey)
	httpContext.Proxy().URI().SetPath(i.path)
	body, _ = json.Marshal(chatRequest)
	httpContext.Proxy().Body().SetRaw("application/json", body)
	if i.balanceHandler != nil {
		ctx.SetBalance(i.balanceHandler)
	}

	return nil
}

func (i *ImageEdit) ResponseConvert(ctx eoscContext.EoContext) error {
	return nil
}
