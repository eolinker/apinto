package main

import (
	"os"

	"github.com/eolinker/goku/professions"

	process_master "github.com/eolinker/eosc/process-master"

	"github.com/eolinker/eosc/log"
	"github.com/eolinker/eosc/pidfile"
)

func ProcessMaster() {
	process_master.InitLogTransport()
	file, err := pidfile.New()
	if err != nil {
		log.Errorf("the process-master is running:%v by:%d", err, os.Getpid())
		return
	}
	master := process_master.NewMasterHandle(file)
	if err := master.Start(NewMasterHandler()); err != nil {
		master.Close()
		log.Errorf("process-master[%d] start faild:%v", os.Getpid(), err)
		return
	}

	master.Wait()
}

func NewMasterHandler() *process_master.MasterHandler {
	return &process_master.MasterHandler{
		Professions: professions.NewProfessions("profession.yml"),
	}
}
