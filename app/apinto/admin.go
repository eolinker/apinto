package main

import (
	"github.com/eolinker/eosc/log"
	process_admin "github.com/eolinker/eosc/process-admin"
)

func ProcessAdmin() {
	log.Debug("start admin")
	registerInnerExtenders()
	process_admin.Process()
}
