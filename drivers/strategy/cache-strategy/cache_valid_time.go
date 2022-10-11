package cache_strategy

import (
	"fmt"
	"github.com/eolinker/eosc"
	"reflect"
)

var (
	_ eosc.IWorker        = (*CacheValidTime)(nil)
	_ eosc.IWorkerDestroy = (*CacheValidTime)(nil)
)

type CacheValidTime struct {
	id        string
	name      string
	handler   *CacheValidTimeHandler
	config    *Config
	isRunning int
}

func (l *CacheValidTime) Destroy() error {
	controller.Del(l.id)
	return nil
}

func (l *CacheValidTime) Id() string {
	return l.id
}

func (l *CacheValidTime) Start() error {
	if l.isRunning == 0 {
		l.isRunning = 1
		actuatorSet.Set(l.id, l.handler)
	}

	return nil
}

func (l *CacheValidTime) Reset(v interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	conf, ok := v.(*Config)
	if !ok {
		return eosc.ErrorConfigIsNil
	}
	if conf.Priority > 999 || conf.Priority < 1 {
		return fmt.Errorf("priority value %d not allow ", conf.Priority)
	}
	if conf.ValidTime < 1 {
		return fmt.Errorf("validTime value %d not allow ", conf.ValidTime)
	}

	confCore := conf
	if reflect.DeepEqual(l.config, confCore) {
		return nil
	}
	handler, err := NewCacheValidTimeHandler(confCore)
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

func (l *CacheValidTime) Stop() error {
	if l.isRunning != 0 {
		l.isRunning = 0
		actuatorSet.Del(l.id)
	}

	return nil
}

func (l *CacheValidTime) CheckSkill(skill string) bool {
	return false
}
