package oauth2

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/router"
)

const (
	Name = "oauth2"
)

func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(Name, NewFactory())
}

func NewFactory() eosc.IExtenderDriverFactory {
	h := NewTokenHandler()
	router.SetPath("oauth2_token_handler", "/oauth_tokens/", h)
	return drivers.NewFactory[Config](Create)
}
