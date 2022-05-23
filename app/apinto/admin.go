package main

import process_admin "github.com/eolinker/eosc/process-admin"

func ProcessAdmin() {
	registerInnerExtenders()
	process_admin.Process()
}
