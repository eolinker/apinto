package upstream_http

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/eolinker/goku-eosc/upstream"

	"github.com/eolinker/eosc"
	"github.com/eolinker/goku-eosc/discovery"

	"github.com/eolinker/goku-eosc/service"

	"github.com/eolinker/goku-eosc/upstream/balance"

	http_proxy "github.com/eolinker/goku-eosc/http-proxy"

	http_context "github.com/eolinker/eosc/node/http-context"

	"github.com/eolinker/eosc/utils"
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
	balanceHandler balance.IBalanceHandler
}

func (h *httpUpstream) Id() string {
	return h.id
}

func (h *httpUpstream) Start() error {
	handler, err := balance.GetDriver(h.balanceType, h.app)
	if err != nil {
		return err
	}
	h.balanceHandler = handler
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
			handler, err := balance.GetDriver(h.balanceType, h.app)
			if err != nil {
				return err
			}
			h.balanceHandler = handler
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
	var response *http.Response
	var err error
	path := utils.TrimPrefixAll(ctx.ProxyRequest.TargetURL(), "/")
	node, err := h.balanceHandler.Next()
	if err != nil {
		return nil, err
	}
	for doTrice := serviceDetail.Retry() + 1; doTrice > 0; doTrice-- {

		u := fmt.Sprintf("%s://%s/%s", h.scheme, node.Addr(), path)
		response, err = http_proxy.DoRequest(ctx, u, serviceDetail.Timeout())

		if err != nil {
			if response == nil {
				node.Down()
			}
			h.app.NodeError(node.Id())
			node, err = h.balanceHandler.Next()
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

func GetType() reflect.Type {
	return reflect.TypeOf((*httpUpstream)(nil))
}
