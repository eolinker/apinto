package upstream_http

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/eolinker/goku-eosc/upstream"

	"github.com/eolinker/eosc"
	"github.com/eolinker/goku-eosc/discovery"

	"github.com/eolinker/goku-eosc/service"

	"github.com/eolinker/goku-eosc/upstream/balance"

	http_proxy "github.com/eolinker/goku-eosc/node/http-proxy"

	http_context "github.com/eolinker/goku-eosc/node/http-context"

	"github.com/eolinker/goku-eosc/utils"
)

//Http org
type httpUpstream struct {
	id             string
	name           string
	driver         string
	desc           string
	scheme         string
	balanceType    string
	app            discovery.IApp
	balanceFactory balance.IBalanceFactory
}

func (h *httpUpstream) Id() string {
	return h.id
}

func (h *httpUpstream) Start() error {
	return nil
}

func (h *httpUpstream) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
	cfg, ok := conf.(*Config)
	if !ok {
		return errors.New(fmt.Sprintf(ErrorStructType, eosc.TypeNameOf(conf), eosc.TypeNameOf((*Config)(nil))))
	}
	if factory, has := workers[cfg.Discovery]; has {
		f, ok := factory.(discovery.IDiscovery)
		if ok {
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
			h.balanceFactory, err = balance.GetFactory(h.balanceType)
			if err != nil {
				return err
			}
			return nil
		}
	}
	return errors.New("fail to create upstream worker")
}

func (h *httpUpstream) Stop() error {
	h.app.Close()
	return nil
}

func (h *httpUpstream) CheckSkill(skill string) bool {
	return upstream.CheckSkill(skill)
}

//send 请求发送，忽略重试
func (h *httpUpstream) Send(ctx *http_context.Context, serviceDetail service.IServiceDetail) (*http.Response, error) {
	handler, err := h.balanceFactory.Create(h.app)
	if err != nil {
		return nil, err
	}
	var response *http.Response
	path := utils.TrimPrefixAll(ctx.ProxyRequest.TargetURL(), "/")

	node, err := handler.Next()
	if err != nil {
		return nil, err
	}
	for doTrice := serviceDetail.Retry() + 1; doTrice > 0; doTrice-- {
		fmt.Println("addr is:", node.Addr())
		u := fmt.Sprintf("%s://%s/%s", h.scheme, node.Addr(), path)
		response, err = http_proxy.DoRequest(ctx, u, serviceDetail.Timeout())

		if err != nil {
			if response == nil {
				node.Down()
			}
			h.app.NodeError(node.ID())
			node, err = handler.Next()
			if err != nil {
				return nil, err
			}
			continue
		} else {
			return response, err
		}
	}

	return response, err
}
