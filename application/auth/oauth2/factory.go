package oauth2

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/eolinker/eosc/utils/schema"

	"github.com/eolinker/apinto/application"
	"github.com/eolinker/apinto/application/auth"
)

var _ auth.IAuthFactory = (*factory)(nil)

var driverName = "oauth2"

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
		"oauth2",
		"oauth2_auth",
	}
}

func (f *factory) PreRouters() []*auth.PreRouter {
	return []*auth.PreRouter{
		{
			ID:         "/oauth2/token",
			PreHandler: NewHandler(NewTokenHandler()),
			Path:       "/oauth2/token",
			Method:     []string{http.MethodPost},
		},
		{
			ID:         "/oauth2/authorize",
			PreHandler: NewHandler(NewAuthorizeHandler()),
			Path:       "/oauth2/authorize",
			Method:     []string{http.MethodPost},
		},
	}
}

func (f *factory) Create(tokenName string, position string, rule interface{}) (application.IAuth, error) {
	a := &oauth2{
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
	return &factory{
		configType: typ,
		render:     render,
		userType:   reflect.TypeOf((*User)(nil)),
	}
}

func toId(tokenName, position string) string {
	return fmt.Sprintf("%s@%s@%s", tokenName, position, driverName)
}
