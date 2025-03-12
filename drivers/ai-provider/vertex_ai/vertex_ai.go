package vertex_ai

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	ai_convert "github.com/eolinker/apinto/ai-convert"
	"github.com/eolinker/eosc/common/bean"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
	"golang.org/x/oauth2"
	"net/url"
	"strings"

	"golang.org/x/oauth2/google"

	dns "google.golang.org/api/dns/v2"
)

var (
	accessConfigManager ai_convert.IModelAccessConfigManager
	scopes              = []string{
		dns.CloudPlatformReadOnlyScope,
		dns.CloudPlatformScope,
	}
)

func init() {
	ai_convert.RegisterConverterCreateFunc("vertex_ai", Create)
	bean.Autowired(&accessConfigManager)
}

type Config struct {
	ProjectID         string `json:"vertex_project_id"`
	Location          string `json:"vertex_location"`
	ServiceAccountKey string `json:"vertex_service_account_key"`
	Base              string `json:"vertex_api_base"`
}

func checkConfig(conf *Config) error {
	// Check if the APIKey is provided. It is a required field.
	if conf.ProjectID == "" {
		return fmt.Errorf("project_id is required")
	}
	if conf.Location == "" {
		return fmt.Errorf("location is required")
	}
	if conf.ServiceAccountKey == "" {
		return fmt.Errorf("service_account_key is required")
	}
	serviceAccountKey, err := base64.StdEncoding.DecodeString(conf.ServiceAccountKey)
	_, err = google.JWTConfigFromJSON(serviceAccountKey)
	if err != nil {
		return err
	}
	if conf.Base != "" {
		tmpBase := ""
		tmpBase = strings.ReplaceAll(conf.Base, "{LOCATION}", conf.Location)
		tmpBase = strings.ReplaceAll(tmpBase, "{PROJECT_ID}", conf.ProjectID)
		u, err := url.Parse(tmpBase)
		if err != nil {
			// Return an error if the Base URL cannot be parsed.
			return fmt.Errorf("base url is invalid")
		}
		// Ensure the parsed URL contains both a scheme and a host.
		if u.Scheme == "" || u.Host == "" {
			return fmt.Errorf("base url is invalid")
		}
	}

	return nil
}

func Create(cfg string) (ai_convert.IConverter, error) {
	var conf Config
	err := json.Unmarshal([]byte(cfg), &conf)
	if err != nil {
		return nil, err
	}
	err = checkConfig(&conf)
	if err != nil {
		return nil, err
	}

	return NewChat(conf)
}

type Chat struct {
	token  *oauth2.Token
	config Config
}

func NewChat(config Config) (ai_convert.IChildConverter, error) {
	jwtData, err := base64.StdEncoding.DecodeString(config.ServiceAccountKey)
	if err != nil {
		return nil, err
	}
	token, err := newToken(context.Background(), jwtData)
	if err != nil {
		return nil, err
	}
	return &Chat{
		token:  token,
		config: config,
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
	provider := ai_convert.GetAIProvider(ctx)
	model := ai_convert.GetAIModel(ctx)
	modelCfg, has := accessConfigManager.Get(fmt.Sprintf("%s$%s", provider, model))
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

	openAIConvert, err := ai_convert.NewOpenAIConvert(c.token.AccessToken, base, 0, nil, nil)
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
	case 429:
		switch resp.Error.Type {
		case "insufficient_quota":
			// Handle the insufficient quota error.
			ai_convert.SetAIStatusQuotaExhausted(ctx)
		case "rate_limit_error":
			// Handle the rate limit error.
			ai_convert.SetAIStatusExceeded(ctx)
		}
	case 401:
		// 过期和无效的API密钥
		ai_convert.SetAIStatusInvalid(ctx)
	}
}
