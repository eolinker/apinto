package openid_connect_jwt

import (
	"fmt"
	"reflect"

	"github.com/eolinker/eosc/router"

	"github.com/eolinker/eosc/utils/schema"

	"github.com/eolinker/apinto/application"
	"github.com/eolinker/apinto/application/auth"
)

var _ auth.IAuthFactory = (*factory)(nil)

var driverName = "openid-connect-jwt"

// Register 注册auth驱动工厂
func Register() {
	auth.FactoryRegister(driverName, NewFactory())
}

type factory struct {
	configType reflect.Type
	render     *schema.Schema
	userType   reflect.Type
}

func (f *factory) Render() interface{} {
	return f.render
}

func (f *factory) ConfigType() reflect.Type {
	return f.configType
}

func (f *factory) UserType() reflect.Type {
	return f.userType
}

func (f *factory) Alias() []string {
	return []string{
		"openid-connect-jwt",
	}
}

func (f *factory) PreRouters() []*auth.PreRouter {
	return []*auth.PreRouter{}
}

func (f *factory) Create(tokenName string, position string, rule interface{}) (application.IAuth, error) {
	a := &jwt{
		id:        toId(tokenName, position),
		tokenName: tokenName,
		position:  position,
		users:     application.NewUserManager(),
	}
	return a, nil
}

// NewFactory 生成一个 auth_apiKey工厂
func NewFactory() auth.IAuthFactory {
	typ := reflect.TypeOf((*Config)(nil))
	render, _ := schema.Generate(typ, nil)
	h := newIssuerHandler("/openid-connect/issuers")
	j := newJwkHandler("/openid-connect/jwks")
	router.SetPath("openid-connect-jwt-issuer", h.prefix, h)
	router.SetPath("openid-connect-jwt-jwk", j.prefix, j)
	return &factory{
		configType: typ,
		render:     render,
		userType:   reflect.TypeOf((*User)(nil)),
	}
}

func toId(tokenName, position string) string {
	return fmt.Sprintf("%s@%s@%s", tokenName, position, driverName)
}
