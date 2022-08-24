package jwt

import (
	"github.com/eolinker/apinto/application"
	"github.com/eolinker/apinto/application/auth"
	"github.com/eolinker/eosc/variable"
	"reflect"
)

var _ auth.IAuthFactory = (*factory)(nil)

var driverName = "jwt"

//Register 注册auth驱动工厂
func Register() {
	auth.Register(driverName, NewFactory())
}

type factory struct {
}

func (f *factory) Create(tokenName string, position string, rule interface{}) (application.IAuth, error) {
	cfg := &Config{}
	_, err := variable.RecurseReflect(reflect.ValueOf(rule), reflect.ValueOf(cfg), nil)
	if err != nil {
		return nil, err
	}
	id, err := cfg.ToID()
	if err != nil {
		return nil, err
	}
	a := &jwt{
		id:        id,
		tokenName: tokenName,
		position:  position,
		cfg:       cfg,
		users:     application.NewUserManager(getUser),
	}
	return a, nil
}

//NewFactory 生成一个 auth_apiKey工厂
func NewFactory() auth.IAuthFactory {
	return &factory{}
}
