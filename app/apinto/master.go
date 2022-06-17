package main

import (
	"github.com/eolinker/apinto/professions"
	process_master "github.com/eolinker/eosc/process-master"
)

func ProcessMaster() {
	handler := &process_master.MasterHandler{
		InitProfession: professions.ApintoProfession,
	}
	process_master.ProcessDo(handler)
}
