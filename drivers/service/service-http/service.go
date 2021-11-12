package service_http

import (
	"errors"
	"fmt"

	"time"

	"github.com/eolinker/eosc"
	"github.com/eolinker/goku/upstream"

	"github.com/eolinker/goku/service"
)

var (
	ErrorStructType   = errors.New("error struct type")
	ErrorNeedUpstream = errors.New("need upstream")
)

type serviceWorker struct {
	id          string
	name        string
	desc        string
	driver      string
	timeout     time.Duration
	retry       int
	scheme      string
	proxyMethod string

	upstream upstream.IUpstreamCreate
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
		return fmt.Errorf("need %s,now %s", eosc.TypeNameOf((*Config)(nil)), eosc.TypeNameOf(conf))
	}
	data.rebuild()
	if data.Upstream == "" && data.UpstreamAnonymous == "" {
		return ErrorNeedUpstream
	}

	upstreamWork, has := workers[data.Upstream]
	if !has {
		upstreamWork = defaultDiscovery.GetApp(data.UpstreamAnonymous)
	}

	//
	s.desc = data.Desc
	s.timeout = time.Duration(data.Timeout) * time.Millisecond

	s.retry = data.Retry
	s.scheme = data.Scheme
	s.proxyMethod = data.ProxyMethod
	s.upstream = nil
	if worker, has := workers[data.Upstream]; has {
		u, ok := worker.(upstream.IUpstreamCreate)
		if ok {
			s.upstream = u
			return nil
		}
	} else {
		s.proxyAddr = string(data.Upstream)
	}

	return nil

}

func (s *serviceWorker) Stop() error {
	return nil
}

//CheckSkill 检查目标能力是否存在
func (s *serviceWorker) CheckSkill(skill string) bool {
	return service.CheckSkill(skill)
}
