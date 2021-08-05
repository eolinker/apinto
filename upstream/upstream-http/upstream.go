package upstream_http

import (
	"errors"
	"fmt"

	"github.com/valyala/fasthttp"

	"github.com/eolinker/goku-eosc/upstream"

	"github.com/eolinker/eosc"
	"github.com/eolinker/goku-eosc/discovery"

	"github.com/eolinker/goku-eosc/service"

	"github.com/eolinker/goku-eosc/upstream/balance"

	http_proxy "github.com/eolinker/goku-eosc/node/http-proxy"

	http_context "github.com/eolinker/goku-eosc/node/http-context"

	"github.com/eolinker/goku-eosc/utils"
)

var (
	errorScheme       = errors.New("error scheme.only support http or https")
	ErrorStructType   = errors.New("error struct type")
	errorCreateWorker = errors.New("fail to create upstream worker")
)

//Http org
type httpUpstream struct {
	id          string
	name        string
	driver      string
	desc        string
	scheme      string
	balanceType string
	app         discovery.IApp
	handler     balance.IBalanceHandler
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
		return fmt.Errorf("need %s,now %s:%w", eosc.TypeNameOf((*Config)(nil)), eosc.TypeNameOf(conf), ErrorStructType)
	}
	if factory, has := workers[cfg.Discovery]; has {
		f, ok := factory.(discovery.IDiscovery)
		if ok {
			if cfg.Scheme != "http" && cfg.Scheme != "https" {
				return errorScheme
			}
			app, err := f.GetApp(cfg.Config)
			if err != nil {
				return err
			}
			h.desc = cfg.Desc
			h.scheme = cfg.Scheme
			h.balanceType = cfg.Type
			h.app = app

			if err != nil {
				return err
			}
			f, err := balance.GetFactory(h.balanceType)
			if err != nil {
				return err
			}
			h.handler, err = f.Create(h.app)
			if err != nil {
				return err
			}
			return nil
		}
	}
	return errorCreateWorker
}

//Stop 停止http_proxy负载，并关闭相应的app
func (h *httpUpstream) Stop() error {
	h.app.Close()
	return nil
}

//CheckSkill 检查目标能力是否存在
func (h *httpUpstream) CheckSkill(skill string) bool {
	return upstream.CheckSkill(skill)
}

//Send 请求发送，忽略重试
func (h *httpUpstream) Send(ctx *http_context.Context, serviceDetail service.IServiceDetail) (*fasthttp.Response, error) {
	var response *fasthttp.Response
	var err error

	path := utils.TrimPrefixAll(ctx.ProxyRequest.TargetURL(), "/")

	for doTrice := serviceDetail.Retry() + 1; doTrice > 0; doTrice-- {
		var node discovery.INode
		node, err = h.handler.Next()
		if err != nil {
			return nil, err
		}
		scheme := node.Scheme()
		if scheme != "http" && scheme != "https" {
			scheme = h.scheme
		}
		u := fmt.Sprintf("%s://%s/%s", scheme, node.Addr(), path)
		response, err = http_proxy.DoRequest(ctx, u, serviceDetail.Timeout())

		if err != nil {
			if response == nil {
				node.Down()
			}
			//处理不可用节点
			h.app.NodeError(node.ID())
			continue
		} else {
			return response, nil
		}
	}

	return response, err
}
