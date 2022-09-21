package aksk

import (
	"fmt"
	"reflect"

	"github.com/eolinker/eosc/utils/schema"

	"github.com/eolinker/apinto/application"
	"github.com/eolinker/apinto/application/auth"
)

var _ auth.IAuthFactory = (*factory)(nil)

var driverName = "aksk"

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
		"ak/sk",
		"aksk",
	}
}

func (f *factory) Create(tokenName string, position string, rule interface{}) (application.IAuth, error) {
	a := &aksk{
		id:        toId(tokenName, position),
		tokenName: tokenName,
		position:  position,
		users:     application.NewUserManager(),
	}
	return a, nil
}

//NewFactory 生成一个 auth_apiKey工厂
func NewFactory() auth.IAuthFactory {
	typ := reflect.TypeOf((*Config)(nil))
	render, _ := schema.Generate(typ, nil)

	return &factory{configType: typ, render: render}
}

func toId(tokenName, position string) string {
	return fmt.Sprintf("%s@%s@%s", tokenName, position, driverName)
}
