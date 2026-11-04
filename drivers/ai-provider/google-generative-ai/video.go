package google

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
	"time"

	ai_convert "github.com/eolinker/apinto/ai-convert"
	context_label "github.com/eolinker/apinto/common/context-label"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"golang.org/x/oauth2"
)

const (
	videoInteractionsPath = "/interactions"
)

func init() {
	driverCreate.Set(ai_convert.ModelTypeVideoGeneration, func(mt ai_convert.ModelType, c *Config) (ai_convert.IConverterDriver, error) {
		return NewVideoTaskCommit(provider, c.ServiceAccountKey, c.BaseUrl, mt, 10*time.Minute)
	})

}

var _ ai_convert.IConverterDriver = (*Video)(nil)

func NewVideoTaskCommit(provider string, apikey string, baseUrl string, modelType ai_convert.ModelType, timeout time.Duration) (ai_convert.IConverterDriver, error) {
	c := &Video{
		provider:  provider,
		saKey:     apikey,
		modelType: modelType,
	}
	if baseUrl == "" {
		baseUrl = generativeAIBase
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
	trimmedPath := strings.TrimSuffix(u.Path, "/")
	if trimmedPath == "" {
		c.path = "/v1beta1" + videoInteractionsPath
	} else if strings.HasSuffix(trimmedPath, videoInteractionsPath) {
		c.path = trimmedPath
	} else {
		c.path = fmt.Sprintf("%s%s", trimmedPath, videoInteractionsPath)
	}
	return c, nil
}

type Video struct {
	provider       string
	path           string
	saKey          string
	token          *oauth2.Token
	balanceHandler eocontext.BalanceHandler
	modelType      ai_convert.ModelType
}

func (c *Video) Provider() string {
	return c.provider
}

func (c *Video) ModelType() ai_convert.ModelType {
	return c.modelType
}

func (c *Video) RequestConvert(ctx eocontext.EoContext, extender map[string]interface{}) error {
	if !c.token.Valid() {
		jwtData, err := base64.StdEncoding.DecodeString(c.saKey)
		if err != nil {
			return err
		}
		token, err := newToken(context.Background(), jwtData)
		if err != nil {
			return err
		}
		c.token = token
	}
	if c.token == nil {
		return fmt.Errorf("token is nil")
	}
	httpContext, err := http_service.Assert(ctx)
	if err != nil {
		return err
	}

	httpContext.Proxy().Header().SetHeader("Authorization", "Bearer "+c.token.AccessToken)
	httpContext.Proxy().URI().SetPath(c.path)
	if c.balanceHandler != nil {
		ctx.SetBalance(c.balanceHandler)
	}
	context_label.SetBillingMode(ctx, context_label.BillingModeImmediate)
	return nil
}

func (c *Video) ResponseConvert(ctx eocontext.EoContext) error {
	httpContext, err := http_service.Assert(ctx)
	if err != nil {
		return err
	}
	ensureFailure(httpContext)
	return nil
}
