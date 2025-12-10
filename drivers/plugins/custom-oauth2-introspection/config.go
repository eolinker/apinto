package custom_oauth2_introspection

import (
	"fmt"
	"github.com/ohler55/ojg/jp"
	"net/url"
	"strings"
)

const (
	positionHeader = "header"
	positionQuery  = "query"
	positionBody   = "body"
)

const (
	redisKeyPrefix = "apinto:custom-oauth2-introspection"
)

const (
	FormData = "form-data"
	Json     = "json"
)

type Config struct {
	Endpoint             string                `json:"endpoint" description:"认证服务地址"`
	IntrospectionRequest *IntrospectionRequest `json:"introspection_request" description:"认证请求参数配置"`
	TokenPosition        string                `json:"token_position" description:"Token放置位置" default:"header"`
	TokenName            string                `json:"token_name" description:"Token名称" default:"access_token"`
	TTL                  int                   `json:"ttl" description:"缓存时间（秒）"`
}

type IntrospectionRequest struct {
	Method          string            `json:"method" description:"请求方法" default:"POST"`
	Header          map[string]string `json:"header"`
	Body            map[string]string `json:"body"`
	Query           map[string]string `json:"query"`
	ContentType     string            `json:"content_type"`
	Retry           int               `json:"retry" description:"超时重试次数" default:"3"`
	ExtractResponse *ExtractResponse  `json:"extract_response"`
}

func (i *IntrospectionRequest) Check() error {
	if i.ContentType == "" {
		i.ContentType = FormData
	}

	if i.Method == "" {
		i.Method = "POST"
	}

	if i.Retry < 0 {
		i.Retry = 3
	}

	if i.ExtractResponse == nil {
		return fmt.Errorf("extract_response is required")
	}

	if err := i.ExtractResponse.Check(); err != nil {
		return err
	}

	return nil
}

type ExtractResponse struct {
	AccessToken *ExtractParam `json:"access_token"`
	ExpiredIn   *ExtractParam `json:"expired_in"`
}

func (e *ExtractResponse) Check() error {
	if e.AccessToken == nil {
		return fmt.Errorf("access_token is required")
	}
	if err := e.AccessToken.Check(); err != nil {
		return fmt.Errorf("access_token is invalid, %w", err)
	}

	if e.ExpiredIn == nil {
		return fmt.Errorf("expired_in is required")
	}
	if err := e.ExpiredIn.Check(); err != nil {
		return fmt.Errorf("expired_in is invalid, %w", err)
	}
	return nil
}

type ExtractParam struct {
	Key      string  `json:"key"`
	Position string  `json:"position" description:"参数位置" enum:"header,query,body"  default:"header"`
	expr     jp.Expr `json:"-"`
}

func (e *ExtractParam) Check() error {
	if e.Key == "" {
		return fmt.Errorf("key is required")
	}
	if e.Position == "" {
		e.Position = positionHeader
	}
	if e.Position != positionHeader && e.Position != positionBody {
		return fmt.Errorf("position is invalid,key:%s", e.Key)
	}
	if e.Position == positionBody {
		if !strings.HasPrefix(e.Key, "$.") {
			e.Key = fmt.Sprintf("$.%s", e.Key)
		}
		expr, err := jp.Parse([]byte(e.Key))
		if err != nil {
			return err
		}
		e.expr = expr
	}
	return nil
}

func Check(conf *Config) error {
	if conf.Endpoint == "" {
		return fmt.Errorf("endpoint is required")
	}

	err := CheckURL(conf.Endpoint)
	if err != nil {
		return err
	}
	if conf.IntrospectionRequest == nil {
		return fmt.Errorf("introspection_request is required")
	}

	if err := conf.IntrospectionRequest.Check(); err != nil {
		return err
	}
	if conf.TokenPosition == "" {
		conf.TokenPosition = positionHeader
	}
	if conf.TokenPosition != positionHeader && conf.TokenPosition != positionBody && conf.TokenPosition != positionQuery {
		return fmt.Errorf("token_position is invalid")
	}
	if conf.TokenName == "" {
		conf.TokenName = "access_token"
	}

	if conf.TTL <= 0 {
		conf.TTL = 600
	}
	return nil
}

func CheckURL(endpoint string) error {
	u, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("endpoint is invalid: %w", err)
	}
	if u.Scheme == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return fmt.Errorf("scheme is invalid: %s", endpoint)
	}
	if u.Host == "" {
		return fmt.Errorf("host is invalid: %s", endpoint)
	}
	return nil
}
