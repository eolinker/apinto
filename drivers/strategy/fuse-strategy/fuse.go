package fuse_strategy

import (
	"fmt"
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"reflect"
)

var (
	_ eosc.IWorker        = (*Fuse)(nil)
	_ eosc.IWorkerDestroy = (*Fuse)(nil)
)

type Fuse struct {
	drivers.WorkerBase
	handler   *FuseHandler
	config    *Config
	isRunning int
}

func (l *Fuse) Destroy() error {
	controller.Del(l.Id())
	return nil
}

func (l *Fuse) Start() error {
	if l.isRunning == 0 {
		l.isRunning = 1
		actuatorSet.Set(l.Id(), l.handler)
	}

	return nil
}

func (l *Fuse) Reset(v interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	conf, ok := v.(*Config)
	if !ok {
		return eosc.ErrorConfigIsNil
	}
	if conf.Priority > 999 || conf.Priority < 1 {
		return fmt.Errorf("priority value %d not allow ", conf.Priority)
	}

	confCore := conf
	if reflect.DeepEqual(l.config, confCore) {
		return nil
	}
	handler, err := NewFuseHandler(confCore)
	if err != nil {
		return err
	}
	l.config = confCore
	l.handler = handler
	if l.isRunning != 0 {
		actuatorSet.Set(l.Id(), l.handler)
	}
	return nil
}

func (l *Fuse) Stop() error {
	if l.isRunning != 0 {
		l.isRunning = 0
		actuatorSet.Del(l.Id())
	}

	return nil
}

func (l *Fuse) CheckSkill(skill string) bool {
	return false
}
