package google

import (
	"fmt"
	ai_convert "github.com/eolinker/apinto/ai-convert"
	context_label "github.com/eolinker/apinto/utils/context-label"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"net/url"
	"strings"
	"time"
)

var (
	generativeAIBase = "https://generativelanguage.googleapis.com/v1beta/models"
	generativeAIPath = "/v1beta/models"
	vertexAIBase     = "https://aiplatform.googleapis.com/v1/publishers/google/models"
	vertexAIPath     = "/v1/publishers/google/models"
)

func init() {
	driverCreate.Set(ai_convert.ModelTypeChat, func(mt ai_convert.ModelType, c *Config) (ai_convert.IConverterDriver, error) {
		return NewChat(provider, c.APIKey, c.BaseUrl, c.Type, mt, 0)
	})
}

func NewChat(provider string, apikey string, baseUrl string, orgType string, modelType ai_convert.ModelType, timeout time.Duration) (ai_convert.IConverterDriver, error) {
	c := &Chat{
		provider:  provider,
		apikey:    apikey,
		modelType: modelType,
	}
	orgBase := generativeAIBase
	orgPath := generativeAIPath
	if orgType == TypeVertexAI {
		orgPath = vertexAIPath
		orgBase = vertexAIBase
	}
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
		c.path = orgPath
	} else {
		c.path = u.Path
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
	context_label.SetBillingMode(ctx, context_label.BillingModeImmediate)
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
		path = fmt.Sprintf("%s:%s", model, requestPath[index+1:])
	}
	httpContext.Proxy().URI().SetPath(path)
	if c.balanceHandler != nil {
		ctx.SetBalance(c.balanceHandler)
	}
	httpContext.Proxy().SetStreamBodyParse(StreamBodyParse)

	return nil
}

func (c *Chat) ResponseConvert(ctx eocontext.EoContext) error {
	return nil
}
