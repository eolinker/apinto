package service

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	session_keep "github.com/eolinker/apinto/upstream/session-keep"

	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	"github.com/eolinker/eosc/log"
	"github.com/eolinker/eosc/utils/config"

	"github.com/eolinker/apinto/discovery"
	"github.com/eolinker/apinto/upstream/balance"
)

var (
	_ eocontext.BalanceHandler      = (*Service)(nil)
	_ eocontext.EoApp               = (*Service)(nil)
	_ eocontext.UpstreamHostHandler = (*Service)(nil)
)

type Service struct {
	eocontext.BalanceHandler
	app discovery.IApp

	scheme  string
	timeout time.Duration

	lastConfig   *Config
	passHost     eocontext.PassHostMod
	upstreamHost string
}

func (s *Service) PassHost() (eocontext.PassHostMod, string) {
	return s.passHost, s.upstreamHost
}

func (s *Service) Nodes() []eocontext.INode {
	all := s.app.Nodes()
	ns := make([]eocontext.INode, 0, len(all))
	for _, n := range all {
		if n.Status() == eocontext.Running {
			ns = append(ns, n)
		}
	}
	return ns
}

func (s *Service) Scheme() string {
	return s.scheme
}

func (s *Service) TimeOut() time.Duration {
	return s.timeout
}

// Reset 重置服务实例的配置
func (s *Service) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) (err error) {
	data, ok := conf.(*Config)
	if !ok {
		return fmt.Errorf("need %s,now %s", config.TypeNameOf((*Config)(nil)), config.TypeNameOf(conf))
	}
	data.rebuild()
	if reflect.DeepEqual(data, s.lastConfig) {
		return nil
	}

	log.Debug("serviceWorker:", data.String())
	if data.Discovery == "" && len(data.Nodes) == 0 {
		return ErrorNeedUpstream
	}
	if data.Discovery != "" && data.Service == "" {
		return ErrorInvalidDiscovery
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

	old := s.app
	s.app = apps

	s.scheme = data.Scheme
	s.timeout = time.Duration(data.Timeout) * time.Millisecond
	balanceHandler := s.BalanceHandler
	if s.lastConfig == nil || s.lastConfig.Balance != data.Balance || s.lastConfig.KeepSession != data.KeepSession {
		balanceFactory, err := balance.GetFactory(data.Balance)
		if err != nil {
			return err
		}

		handler, err := balanceFactory.Create(s, s.scheme, s.timeout)
		if err != nil {
			return err
		}
		if data.KeepSession {
			handler = session_keep.NewSession(handler)
		}
		balanceHandler = handler
	}
	s.BalanceHandler = balanceHandler
	s.passHost = parsePassHost(data.PassHost)
	s.scheme = data.Scheme

	s.upstreamHost = data.UpstreamHost
	s.lastConfig = data
	if old != nil {
		old.Close()
	}
	return nil

}

func parsePassHost(passHost string) eocontext.PassHostMod {
	switch strings.ToLower(passHost) {
	case "pass":
		return eocontext.PassHost
	case "node":
		return eocontext.NodeHost
	case "rewrite":
		return eocontext.ReWriteHost
	}
	return eocontext.PassHost
}

func compareArray[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
