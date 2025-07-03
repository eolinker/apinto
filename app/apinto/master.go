package main

import (
	process_master "github.com/eolinker/eosc/process-master"
)

func ProcessMaster() {
	handler := &process_master.MasterHandler{
		InitProfession: Profession,
	}
	process_master.ProcessDo(handler)
}
