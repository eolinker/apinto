package visit_strategy

import (
	"fmt"
	"github.com/eolinker/eosc"
	"reflect"
)

var (
	_ eosc.IWorker        = (*Visit)(nil)
	_ eosc.IWorkerDestroy = (*Visit)(nil)
)

type Visit struct {
	id        string
	name      string
	handler   *visitHandler
	config    *Config
	isRunning int
}

func (l *Visit) Destroy() error {
	controller.Del(l.id)
	return nil
}

func (l *Visit) Id() string {
	return l.id
}

func (l *Visit) Start() error {
	if l.isRunning == 0 {
		l.isRunning = 1
		actuatorSet.Set(l.id, l.handler)
	}

	return nil
}

func (l *Visit) Reset(v interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	conf, ok := v.(*Config)
	if !ok {
		return eosc.ErrorConfigType
	}
	return l.reset(conf, workers)
}
func (l *Visit) reset(conf *Config, workers map[eosc.RequireId]eosc.IWorker) error {

	if conf.Priority > 999 || conf.Priority < 1 {
		return fmt.Errorf("priority value %d not allow ", conf.Priority)
	}

	confCore := conf
	if reflect.DeepEqual(l.config, confCore) {
		return nil
	}
	handler, err := newVisitHandler(confCore)
	if err != nil {
		return err
	}
	l.config = confCore
	l.handler = handler
	if l.isRunning != 0 {
		actuatorSet.Set(l.id, l.handler)
	}
	return nil
}

func (l *Visit) Stop() error {
	if l.isRunning != 0 {
		l.isRunning = 0
		actuatorSet.Del(l.id)
	}

	return nil
}

func (l *Visit) CheckSkill(skill string) bool {
	return false
}
