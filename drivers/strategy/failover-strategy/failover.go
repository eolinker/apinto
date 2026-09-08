package failover_strategy

import (
	"fmt"
	"reflect"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

var (
	_ eosc.IWorker        = (*Failover)(nil)
	_ eosc.IWorkerDestroy = (*Failover)(nil)
)

type Failover struct {
	drivers.WorkerBase
	handler   *FailoverHandler
	config    *Config
	isRunning int
}

func (f *Failover) Destroy() error {
	controller.Del(f.Id())
	if f.isRunning != 0 {
		f.isRunning = 0
		actuatorSet.Del(f.Id())
	}
	f.handler = nil
	return nil
}

func (f *Failover) Start() error {
	if f.isRunning == 0 {
		f.isRunning = 1
		actuatorSet.Set(f.Id(), f.handler)
	}
	return nil
}

func (f *Failover) Reset(v interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	conf, ok := v.(*Config)
	if !ok {
		return eosc.ErrorConfigIsNil
	}
	if conf.Priority > 999 || conf.Priority < 1 {
		return fmt.Errorf("priority value %d not allowed", conf.Priority)
	}

	confCore := conf
	if reflect.DeepEqual(f.config, confCore) {
		return nil
	}

	handler, err := NewFailoverHandler(confCore)
	if err != nil {
		return err
	}

	f.config = confCore
	f.handler = handler
	if f.isRunning != 0 {
		actuatorSet.Set(f.Id(), f.handler)
	}
	return nil
}

func (f *Failover) Stop() error {
	if f.isRunning != 0 {
		f.isRunning = 0
		actuatorSet.Del(f.Id())
	}
	return nil
}

func (f *Failover) CheckSkill(skill string) bool {
	return false
}
