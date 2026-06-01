package vertex_ai

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	ai_convert "github.com/eolinker/apinto/ai-convert"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"strings"
)

func init() {
	driverCreate.Set(ai_convert.ModelTypeOpenAIChat, NewChat)
}

type Chat struct {
	token     *oauth2.Token
	modelType ai_convert.ModelType
	config    *Config
}

func (c *Chat) Provider() string {
	return provider
}

func (c *Chat) ModelType() ai_convert.ModelType {
	return c.modelType
}

func NewChat(mt ai_convert.ModelType, config *Config) (ai_convert.IConverterDriver, error) {
	jwtData, err := base64.StdEncoding.DecodeString(config.ServiceAccountKey)
	if err != nil {
		return nil, err
	}
	token, err := newToken(context.Background(), jwtData)
	if err != nil {
		return nil, err
	}
	return &Chat{
		token:     token,
		config:    config,
		modelType: mt,
	}, nil
}

func (c *Chat) Endpoint() string {
	return c.config.Base
}

func (c *Chat) RequestConvert(ctx eocontext.EoContext, extender map[string]interface{}) error {
	if !c.token.Valid() {
		jwtData, err := base64.StdEncoding.DecodeString(c.config.ServiceAccountKey)
		if err != nil {
			return err
		}
		token, err := newToken(context.Background(), jwtData)
		if err != nil {
			return err
		}
		c.token = token
	}
	pro := ai_convert.GetAIProvider(ctx)
	model := ai_convert.GetAIModel(ctx)
	modelCfg, has := accessConfigManager.Get(fmt.Sprintf("%s$%s", pro, model))
	base := c.config.Base
	vertexProjectId, vertexLocation := c.config.ProjectID, c.config.Location
	if has {
		if modelCfg.Config()["vertex_project_id"] != "" {
			vertexProjectId = modelCfg.Config()["vertex_project_id"]
		}
		if modelCfg.Config()["vertex_location"] != "" {
			vertexLocation = modelCfg.Config()["vertex_location"]
		}
		if modelCfg.Config()["vertex_model"] != "" {
			ai_convert.SetAIModel(ctx, modelCfg.Config()["vertex_model"])
		}
	}
	base = strings.ReplaceAll(base, "{PROJECT_ID}", vertexProjectId)
	base = strings.ReplaceAll(base, "{LOCATION}", vertexLocation)

	openAIConvert, err := ai_convert.NewOpenAIChat("", c.token.AccessToken, base, c.modelType, 0, nil, nil)
	if err != nil {
		return err
	}

	return openAIConvert.RequestConvert(ctx, extender)
}

func (c *Chat) ResponseConvert(ctx eocontext.EoContext) error {
	return ai_convert.ResponseConvert(ctx, nil, errorCallback)
}

func newToken(ctx context.Context, data []byte) (*oauth2.Token, error) {
	cfg, err := google.JWTConfigFromJSON(data, scopes...)
	if err != nil {
		return nil, err
	}
	return cfg.TokenSource(ctx).Token()
}

func errorCallback(ctx http_service.IHttpContext, body []byte) {
	var resp ai_convert.Response
	err := json.Unmarshal(body, &resp)
	if err != nil {
		log.Errorf("unmarshal body error: %v, body: %s", err, string(body))
		return
	}
	switch ctx.Response().StatusCode() {

	case 400:
		// Handle the bad request error.
		ai_convert.SetAIStatusInvalidRequest(ctx)
	case 403:
		ai_convert.SetAIStatusInvalid(ctx)
	case 429:
		ai_convert.SetAIStatusQuotaExhausted(ctx)
	case 401:
		// 过期和无效的API密钥
		ai_convert.SetAIStatusInvalid(ctx)
	}
}
