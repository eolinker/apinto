package main

import (
	"os"

	"github.com/eolinker/eosc"

	"github.com/eolinker/goku/professions"

	process_master "github.com/eolinker/eosc/process-master"
	"github.com/eolinker/eosc/utils"

	"github.com/eolinker/eosc/log"
	"github.com/eolinker/eosc/pidfile"
)

func ProcessMaster() {
	p, err := NewMasterHandler()
	if err != nil {
		log.Errorf("fail to read procession.yml: %v", err)
		return
	}
	utils.InitLogTransport(eosc.ProcessMaster)
	file, err := pidfile.New()
	if err != nil {
		log.Errorf("the process-master is running:%v by:%d", err, os.Getpid())
		return
	}
	master := process_master.NewMasterHandle(file)

	if err := master.Start(p); err != nil {
		master.Close()
		log.Errorf("process-master[%d] start faild:%v", os.Getpid(), err)
		return
	}

	master.Wait()
}

func NewMasterHandler() (*process_master.MasterHandler, error) {
	p, err := professions.NewProfessions("profession.yml")
	if err != nil {
		return nil, err
	}
	return &process_master.MasterHandler{
		Professions: p,
	}, nil
}
