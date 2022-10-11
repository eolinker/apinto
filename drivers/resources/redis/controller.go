package redis

import (
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/env"
	"github.com/go-redis/redis/v8"
	"reflect"
)

type Controller struct {
	current *_Cacher
	config  Config
}

func (m *Controller) ConfigType() reflect.Type {
	return configType
}

func (m *Controller) Set(conf interface{}) (err error) {
	config, ok := conf.(*Config)
	if ok && config != nil {
		old := m.config
		m.config = *config

		if env.Process() == eosc.ProcessWorker {
			// todo open or close redis
		}
		redis.NewClusterClient().Close()
	}
	return nil
}

func (m *Controller) Get() interface{} {
	return m.config
}

func (m *Controller) Mode() eosc.SettingMode {
	return eosc.SettingModeSingleton
}

func (m *Controller) Check(cfg interface{}) (profession, name, driver, desc string, err error) {
	err = eosc.ErrorUnsupportedKind
	return
}

func (m *Controller) AllWorkers() []string {
	return []string{"redis@setting"}
}

func NewController() *Controller {
	return &Controller{}
}
