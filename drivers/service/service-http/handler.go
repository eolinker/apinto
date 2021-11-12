package service_http

import (
	"fmt"

	http_service "github.com/eolinker/eosc/http-service"
	"github.com/eolinker/eosc/log"
	"github.com/eolinker/goku/filter"
	http_proxy "github.com/eolinker/goku/node/http-proxy"
	"github.com/eolinker/goku/plugin"
	"github.com/eolinker/goku/service"
	"github.com/eolinker/goku/utils"
	"github.com/valyala/fasthttp"
)

type ServiceHandler struct {
	orgPlugin plugin.IPlugin
	executor  filter.IChain

	proxyMethod string
}

func (s *ServiceHandler) DoFilter(ctx http_service.IHttpContext, next http_service.IChain) (err error) {
	// 构造context
	defer func() {
		if e := recover(); e != nil {
			log.Warn(e)
		}
	}()

	if s.proxyMethod != "" {
		ctx.Proxy().SetMethod(s.proxyMethod)
	}

	return nil
}

//Handle 将服务发送到负载
func (s *ServiceHandler) Handle(ctx http_service.IHttpContext, router service.IRouterEndpoint) error {
	ctx.WithValue("router.endpoint", router)
	return s.executor.DoChain(ctx)

}

func (s *ServiceHandler) send(ctx http_service.IHttpContext, serviceDetail service.IServiceDetail, uri string, query string) (*fasthttp.Response, error) {
	if s.upstream == nil {
		var response *fasthttp.Response
		var err error
		request := ctx.ProxyRequest()
		path := utils.TrimPrefixAll(uri, "/")
		for doTrice := serviceDetail.Retry() + 1; doTrice > 0; doTrice-- {
			u := fmt.Sprintf("%s://%s/%s", serviceDetail.Scheme(), serviceDetail.ProxyAddr(), path)
			request.SetRequestURI(u)
			request.URI().SetQueryString(query)
			response, err = http_proxy.DoRequest(request, serviceDetail.Timeout())

			if err != nil {
				continue
			} else {
				return response, err
			}
		}
		return response, err
	} else {
		return s.upstream.Send(ctx, serviceDetail, uri, query)
	}
}
