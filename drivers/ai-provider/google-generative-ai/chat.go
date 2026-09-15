package google

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	ai_convert "github.com/eolinker/apinto/ai-convert"
	context_label2 "github.com/eolinker/apinto/common/context-label"
	"github.com/eolinker/apinto/encoder"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
)

var (
	generativeAIBase = "https://generativelanguage.googleapis.com/v1beta"
	generativeAIPath = "/models"
)

func init() {
	driverCreate.Set(ai_convert.ModelTypeChat, func(mt ai_convert.ModelType, c *Config) (ai_convert.IConverterDriver, error) {
		return NewChat(provider, c.APIKey, c.BaseUrl, mt, 10*time.Minute)
	})
}

func NewChat(provider string, apikey string, baseUrl string, modelType ai_convert.ModelType, timeout time.Duration) (ai_convert.IConverterDriver, error) {
	c := &Chat{
		provider:  provider,
		apikey:    apikey,
		modelType: modelType,
	}
	orgBase := generativeAIBase
	orgPath := generativeAIPath
	if baseUrl == "" {
		baseUrl = orgBase
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
		c.path = "/v1beta" + orgPath
	} else {
		c.path = fmt.Sprintf("%s%s", strings.TrimSuffix(u.Path, "/"), orgPath)
	}
	return c, nil
}

type Chat struct {
	apikey         string
	provider       string
	path           string
	checkErr       ai_convert.CheckError
	errorCallback  func(ctx http_service.IHttpContext, body []byte)
	balanceHandler eocontext.BalanceHandler
	modelType      ai_convert.ModelType
}

func (c *Chat) Provider() string {
	return c.provider
}

func (c *Chat) ModelType() ai_convert.ModelType {
	return c.modelType
}

func (c *Chat) RequestConvert(ctx eocontext.EoContext, extender map[string]interface{}) error {
	context_label2.SetBillingMode(ctx, context_label2.BillingModeImmediate)
	httpContext, err := http_service.Assert(ctx)
	if err != nil {
		return err
	}

	if c.apikey != "" {
		httpContext.Proxy().Header().SetHeader("x-goog-api-key", c.apikey)
	}
	model := ai_convert.GetAIModel(ctx)
	requestPath := httpContext.Request().URI().Path()
	index := strings.Index(requestPath, ":")
	path := fmt.Sprintf("%s:generateContent", model)
	if index > 0 {
		path = fmt.Sprintf("%s/%s:%s", c.path, model, requestPath[index+1:])
	}
	if strings.Contains(requestPath, "generateContent") {
		context_label2.SetDisableStream(ctx, true)
	}
	httpContext.Proxy().URI().SetPath(path)
	if c.balanceHandler != nil {
		ctx.SetBalance(c.balanceHandler)
	}
	//httpContext.Proxy().SetStreamBodyParse(StreamBodyParse)
	context_label2.SetResponseChunkFunc(ctx, func(ctx eocontext.EoContext) ([]byte, error) {
		hCtx, err := http_service.Assert(ctx)
		if err != nil {
			return nil, err
		}
		body := hCtx.Response().GetBody()
		encoding := hCtx.Response().Headers().Get("content-encoding")
		if encoding != "utf-8" && encoding != "" {
			body, err = encoder.ToUTF8(encoding, body)
			if err != nil {
				log.Errorf("[dynamic-billing] failed to convert response body to UTF-8: %v", err)
			}
		}

		httpContext.Proxy().SetStreamBodyParse(StreamBodyParse)
		return body, nil
	})
	return nil
}

func (c *Chat) ResponseConvert(ctx eocontext.EoContext) error {
	httpContext, err := http_service.Assert(ctx)
	if err != nil {
		return err
	}
	ensureFailure(httpContext)
	return nil
}
