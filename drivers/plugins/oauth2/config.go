package oauth2

import (
	"sync"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

const (
	GrantAuthorizationCode = "authorization_code"
	GrantClientCredentials = "client_credentials"
	GrantRefreshToken      = "refresh_token"
)

type Config struct {
	AcceptHttpIfAlreadyTerminated bool `json:"accept_http_if_already_terminated"`
	//Anonymous                     interface{} `json:"anonymous"`
	EnableAuthorizationCode bool `json:"enable_authorization_code"`
	EnableClientCredentials bool `json:"enable_client_credentials"`
	EnableImplicitGrant     bool `json:"enable_implicit_grant"`
	//EnablePasswordGrant           bool        `json:"enable_password_grant"`
	MandatoryScope         bool `json:"mandatory_scope"`
	PersistentRefreshToken bool `json:"persistent_refresh_token"`
	//Pkce                   string      `json:"pkce"`
	ProvisionKey      string   `json:"provision_key"`
	RefreshTokenTtl   int      `json:"refresh_token_ttl"`
	ReuseRefreshToken bool     `json:"reuse_refresh_token"`
	Scopes            []string `json:"scopes"`
	TokenExpiration   int      `json:"token_expiration"`
}

func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	return &executor{
		WorkerBase: drivers.Worker(id, name),
		cfg:        conf,
		once:       sync.Once{},
	}, nil
}
