package redis

import (
	"context"
	"github.com/eolinker/apinto/resources"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/env"
	"github.com/eolinker/eosc/log"
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
func (m *Controller) shutdown() {
	oldClient := m.current
	if oldClient != nil {
		m.current = nil
		resources.ReplaceCacher()
		oldClient.client.Close()
	}
}
func (m *Controller) Set(conf interface{}) (err error) {
	config, ok := conf.(*Config)
	if ok && config != nil {
		old := m.config
		m.config = *config

		if reflect.DeepEqual(old, m.config) {
			return nil
		}

		if len(m.config.Addrs) == 0 {
			oldClient := m.current
			if oldClient != nil {
				resources.ReplaceCacher()
				m.current = nil
				oldClient.client.Close()
			}
			return nil
		}

		client := redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    m.config.Addrs,
			Username: m.config.Username,
			Password: m.config.Password,
		})
		if res, errPing := client.Ping(context.Background()).Result(); errPing != nil {
			log.Info("ping redis:", res, " error:", err)
			client.Close()
			return errPing
		}

		if env.Process() == eosc.ProcessWorker {
			if m.current == nil {
				m.current = newCacher(client)
				resources.ReplaceCacher(m.current)
			} else {
				m.current.client = client
			}
		} else {
			client.Close()
		}
	} else {
		oldClient := m.current
		if oldClient != nil {
			resources.ReplaceCacher()
			m.current = nil
			oldClient.client.Close()
		}
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
