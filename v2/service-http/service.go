package service_http

import (
	"fmt"
	"github.com/eolinker/apinto/discovery"
	"github.com/eolinker/apinto/upstream/balance"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	"github.com/eolinker/eosc/log"
	"github.com/eolinker/eosc/utils/config"
	"reflect"
	"strings"
	"time"
)

type Service struct {
	eocontext.BalanceHandler
	app discovery.IApp

	scheme  string
	timeout time.Duration

	lastConfig *Config
}

func (s *Service) Nodes() []eocontext.INode {
	return s.app.Nodes()
}

func (s *Service) Scheme() string {
	return s.scheme
}

func (s *Service) TimeOut() time.Duration {
	return s.timeout
}

//Reset 重置服务实例的配置
func (s *Service) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {

	data, ok := conf.(*Config)
	if !ok {
		return fmt.Errorf("need %s,now %s", config.TypeNameOf((*Config)(nil)), config.TypeNameOf(conf))
	}
	data.rebuild()
	if reflect.DeepEqual(data, s.lastConfig) {
		return nil
	}
	s.lastConfig = data

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

	s.timeout = time.Duration(data.Timeout) * time.Millisecond
	s.BalanceHandler = balanceHandler
 
	return nil

}
