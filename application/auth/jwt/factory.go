package jwt

import (
	"errors"
	"reflect"

	"github.com/eolinker/eosc/log"

	"github.com/eolinker/eosc/utils/schema"

	"github.com/eolinker/apinto/application"
	"github.com/eolinker/apinto/application/auth"
)

var _ auth.IAuthFactory = (*factory)(nil)

var driverName = "jwt"

//Register 注册auth驱动工厂
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
		"jwt",
	}
}

func (f *factory) Create(tokenName string, position string, rule interface{}) (application.IAuth, error) {
	baseConfig, ok := rule.(*application.BaseConfig)
	if !ok {
		return nil, errors.New("invalid jwt config")
	}
	log.Debug("config type: ", reflect.TypeOf(baseConfig.Config()))
	cfg, ok := baseConfig.Config().(*Rule)
	if !ok {
		return nil, errors.New("invalid jwt config")
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
		users:     application.NewUserManager(),
	}
	return a, nil
}

//NewFactory 生成一个 auth_apiKey工厂
func NewFactory() auth.IAuthFactory {
	typ := reflect.TypeOf((*Config)(nil))
	render, _ := schema.Generate(typ, nil)
	return &factory{
		configType: reflect.TypeOf((*Rule)(nil)),
		render:     render,
		userType:   reflect.TypeOf((*User)(nil)),
	}
}
