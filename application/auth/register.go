package auth

import (
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/setting"
)

func Register(register eosc.IExtenderDriverRegister) {
	setting.RegisterSetting("auth", defaultAuthFactoryRegister)
}
