package upstream_http

import (
	"errors"
	"fmt"
	"github.com/eolinker/eosc/utils/config"
	"strings"
	"time"

	"github.com/eolinker/eosc/log"

	"github.com/eolinker/apinto/upstream"

	"github.com/eolinker/apinto/discovery"
	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/upstream/balance"
)

var (
	errorScheme          = errors.New("error scheme.only support http-service or https")
	ErrorStructType      = errors.New("error struct type")
	errorCreateWorker    = errors.New("fail to create upstream worker")
	ErrorUpstreamNotInit = errors.New("upstream not init")
)
var _ upstream.IUpstream = (*httpUpstream)(nil)

//Http org
type httpUpstream struct {
	upstream  *Upstream
	id        string
	name      string
	desc      string
	lastError error
}

//func (h *httpUpstream) Merge(high map[string]*plugin.Config) map[string]*plugin.Config {
//	if h.upstream == nil {
//		return high
//	}
//	return h.upstream.Merge(high)
//}

func (h *httpUpstream) Create(id string, retry int, time time.Duration) (upstream.IUpstreamHandler, error) {
	if h.upstream == nil {
		return nil, ErrorUpstreamNotInit
	}
	return h.upstream.Create(id, retry, time)
}

//Id 返回worker id
func (h *httpUpstream) Id() string {
	return h.id
}

func (h *httpUpstream) Start() error {
	return nil
}

//Reset 重新设置http_proxy负载的配置
func (h *httpUpstream) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
	cfg, ok := conf.(*Config)
	if !ok || cfg == nil {
		return fmt.Errorf("need %s,now %s:%w", config.TypeNameOf((*Config)(nil)), config.TypeNameOf(conf), ErrorStructType)
	}

	if factory, has := workers[cfg.Discovery]; has {
		discoveryFactory, ok := factory.(discovery.IDiscovery)
		if ok {
			Scheme := strings.ToLower(cfg.Scheme)
			if Scheme != "http" && Scheme != "https" {
				return errorScheme
			}
			balanceFactory, err := balance.GetFactory(cfg.Type)
			if err != nil {
				return err
			}

			app, err := discoveryFactory.GetApp(cfg.Config)
			if err != nil {
				return err
			}
			balanceHandler, err := balanceFactory.Create(app)
			if err != nil {
				return err
			}

			h.desc = cfg.Desc

			if h.upstream == nil {
				h.upstream = NewUpstream(Scheme, app, balanceHandler)
			} else {
				old := h.upstream.app
				h.upstream.Reset(Scheme, app, balanceHandler)
				closeError := old.Close()
				if closeError != nil {

					log.Warn("close app:", closeError)
				}
			}

			return nil
		}
	}
	return errorCreateWorker
}

//Stop 停止http_proxy负载，并关闭相应的app
func (h *httpUpstream) Stop() error {
	if h.upstream != nil {
		h.upstream.app.Close()

		h.upstream.destroy()
		h.upstream = nil
	}

	return nil
}

//CheckSkill 检查目标能力是否存在
func (h *httpUpstream) CheckSkill(skill string) bool {
	return upstream.CheckSkill(skill)
}
