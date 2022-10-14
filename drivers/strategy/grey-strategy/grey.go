package grey_strategy

import (
	"fmt"
	"github.com/eolinker/eosc"
	"reflect"
)

var (
	_ eosc.IWorker        = (*Grey)(nil)
	_ eosc.IWorkerDestroy = (*Grey)(nil)
)

type Grey struct {
	id        string
	name      string
	handler   *GreyHandler
	config    *Config
	isRunning int
}

func (l *Grey) Destroy() error {
	controller.Del(l.id)
	return nil
}

func (l *Grey) Id() string {
	return l.id
}

func (l *Grey) Start() error {
	if l.isRunning == 0 {
		l.isRunning = 1
		actuatorSet.Set(l.id, l.handler)
	}

	return nil
}

func (l *Grey) Reset(v interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
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
	handler, err := NewGreyHandler(confCore)
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

func (l *Grey) Stop() error {
	if l.isRunning != 0 {
		l.isRunning = 0
		actuatorSet.Del(l.id)
	}

	return nil
}

func (l *Grey) CheckSkill(skill string) bool {
	return false
}
