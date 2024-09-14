package data_mask_strategy

import (
	"reflect"

	"github.com/eolinker/eosc"
)

var (
	_ eosc.IWorker        = (*executor)(nil)
	_ eosc.IWorkerDestroy = (*executor)(nil)
)

type executor struct {
	id     string
	name   string
	config *Config
	//handler *handler
}

func (l *executor) Destroy() error {
	controller.Del(l.id)
	return nil
}

func (l *executor) Id() string {
	return l.id
}

func (l *executor) Start() error {
	//if l.isRunning == 0 {
	//	l.isRunning = 1
	//}

	return nil
}

func (l *executor) Reset(v interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	conf, ok := v.(*Config)
	if !ok {
		return eosc.ErrorConfigType
	}
	return l.reset(conf, workers)
}
func (l *executor) reset(conf *Config, workers map[eosc.RequireId]eosc.IWorker) error {

	confCore := conf
	if reflect.DeepEqual(l.config, confCore) {
		return nil
	}

	h, err := newHandler(confCore)
	if err != nil {
		return err
	}
	l.config = confCore
	//l.handler = h
	//if l.isRunning != 0 {
	actuatorSet.Set(l.id, h)
	//}
	return nil
}

func (l *executor) Stop() error {
	//if l.isRunning != 0 {
	//	l.isRunning = 0
	//}
	actuatorSet.Del(l.id)

	return nil
}

func (l *executor) CheckSkill(skill string) bool {
	return false
}
