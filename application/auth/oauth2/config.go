package oauth2

import "github.com/eolinker/apinto/application"

const (
	GrantAuthorizationCode = "authorization_code"
	GrantClientCredentials = "client_credentials"
	GrantRefreshToken      = "refresh_token"
)

type Config struct {
	application.Auth
	Users []*User `json:"users" label:"用户列表"`
}

type User struct {
	Pattern Pattern `json:"pattern" label:"用户信息"`
	application.User
}

type Pattern struct {
	ClientId                      string   `json:"client_id"`
	ClientSecret                  string   `json:"client_secret"`
	ClientType                    string   `json:"client_type"`
	HashSecret                    bool     `json:"hash_secret"`
	RedirectUrls                  []string `json:"redirect_urls" label:"重定向URL"`
	Scopes                        []string `json:"scopes" label:"授权范围"`
	MandatoryScope                bool     `json:"mandatory_scope" label:"强制授权"`
	ProvisionKey                  string   `json:"provision_key" label:"Provision Key"`
	TokenExpiration               int      `json:"token_expiration" label:"令牌过期时间"`
	RefreshTokenTTL               int      `json:"refresh_token_ttl" label:"刷新令牌TTL"`
	EnableAuthorizationCode       bool     `json:"enable_authorization_code"  label:"启用授权码模式"`
	EnableImplicitGrant           bool     `json:"enable_implicit_grant" label:"启用隐式授权模式"`
	EnableClientCredentials       bool     `json:"enable_client_credentials" label:"启用客户端凭证模式"`
	AcceptHttpIfAlreadyTerminated bool     `json:"accept_http_if_already_terminated" label:"如果已终止，则接受HTTP"`
	ReuseRefreshToken             bool     `json:"reuse_refresh_token" label:"重用刷新令牌"`
	PersistentRefreshToken        bool     `json:"persistent_refresh_token" label:"持久刷新令牌"`
}

func (u *User) Username() string {
	return u.Pattern.ClientId
}
