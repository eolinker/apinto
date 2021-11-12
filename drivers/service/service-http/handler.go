package service_http

import (
	"fmt"

	"github.com/eolinker/goku/upstream"

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
	upstream  upstream.IUpstream
}

func (s *ServiceHandler) DoFilter(ctx http_service.IHttpContext, next http_service.IChain) (err error) {
	// 构造context
	defer func() {
		if e := recover(); e != nil {
			log.Warn(e)
		}
		if ctx.StatusCode() == 0 {
			ctx.SetStatus(200, "200")
		}
	}()
	path := s.rewriteURL

	if s.proxyMethod != "" {
		ctx.ProxyRequest().Header.SetMethod(s.proxyMethod)
	}
	body, err := ctx.BodyHandler().RawBody()
	if err != nil {
		ctx.SetBody([]byte(err.Error()))
		ctx.SetStatus(500)
		return err
	}
	ctx.ProxyRequest().SetBody(body)
	var response *fasthttp.Response
	response, err = s.send(ctx, s, path, string(ctx.RequestOrg().URI().QueryString()))
	if err != nil {
		ctx.SetBody([]byte(err.Error()))
		ctx.SetStatus(500)
		return err
	}
	ctx.se(response)
	if next != nil {
		next.DoChain(ctx)
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
