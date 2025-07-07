package main

import "github.com/eolinker/eosc"

func Register(extenderRegister eosc.IExtenderDriverRegister) {
	driverRegister(extenderRegister)
	pluginRegister(extenderRegister)
}
