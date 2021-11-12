package service_http

import (
	"errors"
	"fmt"

	upstream_http "github.com/eolinker/goku/drivers/upstream/upstream-http"
	"github.com/eolinker/goku/plugin"

	"time"

	"github.com/eolinker/eosc"
	"github.com/eolinker/goku/upstream"

	"github.com/eolinker/goku/service"
)

var (
	ErrorStructType      = errors.New("error struct type")
	ErrorNeedUpstream    = errors.New("need upstream")
	ErrorInvalidUpstream = errors.New("not upstream")
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

	upstream upstream.IUpstream

	pluginConfig map[string]*plugin.Config
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
	var upstreamCreate upstream.IUpstream
	if upstreamWork, has := workers[data.Upstream]; has {
		if up, ok := upstreamWork.(upstream.IUpstream); ok {
			upstreamCreate = up
		} else {
			return fmt.Errorf("%s:%w", data.Upstream, ErrorInvalidUpstream)
		}
	} else {
		anonymous, err := defaultDiscovery.GetApp(data.UpstreamAnonymous)
		if err != nil {
			return err
		}
		upstreamCreate = upstream_http.NewUpstream(s.scheme)
	}

	//
	s.desc = data.Desc
	s.timeout = time.Duration(data.Timeout) * time.Millisecond

	s.retry = data.Retry
	s.scheme = data.Scheme
	s.proxyMethod = data.ProxyMethod
	s.upstream = nil

	upstreamWork.(upstream.IUpstream).Create()

	return nil

}

func (s *serviceWorker) Stop() error {
	return nil
}

//CheckSkill 检查目标能力是否存在
func (s *serviceWorker) CheckSkill(skill string) bool {
	return service.CheckSkill(skill)
}
