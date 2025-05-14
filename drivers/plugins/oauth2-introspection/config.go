package oauth2_introspection

import (
	"fmt"
	"net/url"
)

const (
	positionHeader = "header"
	positionQuery  = "query"
	positionBody   = "body"
)

const (
	redisKeyPrefix = "apinto:oauth2-introspection"
)

type Config struct {
	IntrospectionEndpoint  string   `json:"introspection_endpoint"`
	IntrospectionSSLVerify bool     `json:"introspection_ssl_verify" default:"true"`
	ClientID               string   `json:"client_id"`
	ClientSecret           string   `json:"client_secret"`
	TokenHeader            string   `json:"token_header"`
	Scopes                 []string `json:"scopes"`
	TTL                    int      `json:"ttl" default:"600"`
	CustomClaimsForward    []string `json:"custom_claims_forward"`
	ConsumerBy             string   `json:"consumer_by"`
	AllowAnonymous         bool     `json:"allow_anonymous" default:"false"`
	HideCredential         bool     `json:"hide_credential" default:"false"`
}

func Check(conf *Config) error {
	if conf.IntrospectionEndpoint == "" {
		return fmt.Errorf("introspection_endpoint is required")
	}
	u, err := url.Parse(conf.IntrospectionEndpoint)
	if err != nil {
		return fmt.Errorf("introspection_endpoint is invalid: %w", err)
	}
	if u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("introspection_endpoint is invalid: %s", conf.IntrospectionEndpoint)
	}

	if conf.ClientID == "" {
		return fmt.Errorf("client_id is required")
	}

	if conf.ClientSecret == "" {
		return fmt.Errorf("client_secret is required")
	}

	if conf.TokenHeader == "" {
		conf.TokenHeader = "Authorization"
	}

	if conf.ConsumerBy == "" {
		conf.ConsumerBy = "client_id"
	}

	if conf.TTL <= 0 {
		conf.TTL = 600
	}
	return nil
}
