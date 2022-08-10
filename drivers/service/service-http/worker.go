package service_http

import (
	"errors"
	"fmt"
	"github.com/eolinker/apinto/discovery"
	"github.com/eolinker/apinto/service"
	"github.com/eolinker/apinto/upstream/balance"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/log"
	"github.com/eolinker/eosc/utils/config"
	"strings"
	"time"
)

var (
	ErrorNeedUpstream = errors.New("need upstream")

	ErrorInvalidDiscovery = errors.New("invalid Discovery")
)

type serviceWorker struct {
	Service
	id     string
	name   string
	driver string
}

//Id 返回服务实例 worker id
func (s *serviceWorker) Id() string {
	return s.id
}

func (s *serviceWorker) Start() error {
	return nil
}

//Reset 重置服务实例的配置
func (s *serviceWorker) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
	data, ok := conf.(*Config)
	if !ok {
		return fmt.Errorf("need %s,now %s", config.TypeNameOf((*Config)(nil)), config.TypeNameOf(conf))
	}
	data.rebuild()

	log.Debug("serviceWorker:", data.String())
	if data.Discovery == "" && len(data.Nodes) == 0 {
		return ErrorNeedUpstream
	}
	if data.Discovery != "" && data.Service == "" {
		return ErrorInvalidDiscovery
	}
	balanceFactory, err := balance.GetFactory(data.Balance)
	if err != nil {
		return err
	}
	var apps discovery.IApp
	if data.Discovery != "" {
		discoveryWorker, has := workers[data.Discovery]
		if !has {
			return fmt.Errorf("%s:%w", data.Discovery, ErrorInvalidDiscovery)
		}
		ds, ok := discoveryWorker.(discovery.IDiscovery)
		if !ok {
			return fmt.Errorf("%s:%w", data.Discovery, ErrorInvalidDiscovery)
		}
		apps, err = ds.GetApp(data.Service)
		if err != nil {
			return err
		}
	} else {
		apps, err = defaultHttpDiscovery.GetApp(strings.Join(data.Nodes, ";"))
		if err != nil {
			return err
		}
	}
	balanceHandler, err := balanceFactory.Create(apps)
	if err != nil {
		return err
	}

	s.Service.timeout = time.Duration(data.Timeout) * time.Millisecond

	s.Service.retry = data.Retry

	log.Debug("reset service:", data.Plugins)
	s.Service.reset(data.Scheme, apps, balanceHandler, data.Plugins)

	return nil

}

func (s *serviceWorker) Stop() error {
	for _, h := range s.handlers.List() {
		h.Destroy()
	}
	return nil
}

//CheckSkill 检查目标能力是否存在
func (s *serviceWorker) CheckSkill(skill string) bool {
	return service.CheckSkill(skill)
}
