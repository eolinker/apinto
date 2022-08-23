package apikey

import (
	"fmt"
	"github.com/eolinker/apinto/application"
	"github.com/eolinker/apinto/application/auth"
)

var _ auth.IAuthFactory = (*factory)(nil)

var driverName = "apikey"

//Register 注册auth驱动工厂
func Register() {
	auth.Register(driverName, NewFactory())
}

type factory struct {
}

func (f *factory) Create(tokenName string, position string, rule interface{}) (application.IAuth, error) {
	a := &apikey{
		id:        toId(tokenName, position),
		tokenName: tokenName,
		position:  position,
	}
	return a, nil
}

//NewFactory 生成一个 auth_apiKey工厂
func NewFactory() auth.IAuthFactory {
	return &factory{}
}

func toId(tokenName, position string) string {
	return fmt.Sprintf("%s@%s", tokenName, position)
}
