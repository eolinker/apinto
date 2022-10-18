package redis

import (
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/setting"
	"reflect"
)

var (
	singleton  *Controller
	_          eosc.ISetting = singleton
	configType               = reflect.TypeOf(new(Config))
)

func init() {
	singleton = NewController()
}

func Register(register eosc.IExtenderDriverRegister) {
	setting.RegisterSetting("redis", singleton)
}
