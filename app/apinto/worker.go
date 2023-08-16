package main

import (
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/extends"
	process_worker "github.com/eolinker/eosc/process-worker"
)

func ProcessWorker() {
	registerInnerExtenders()
	process_worker.Process()
}
func registerInnerExtenders() {
	extends.AddInnerExtendProject("eolinker.com", "apinto", Register)
}

func Register(extenderRegister eosc.IExtenderDriverRegister) {
	driverRegister(extenderRegister)
	pluginRegister(extenderRegister)
}
