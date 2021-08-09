package service_http

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/valyala/fasthttp"

	http_proxy "github.com/eolinker/goku-eosc/node/http-proxy"
	"github.com/eolinker/goku-eosc/utils"

	"github.com/eolinker/eosc/log"

	"github.com/eolinker/goku-eosc/auth"

	"github.com/eolinker/goku-eosc/router/checker"

	"github.com/eolinker/eosc"
	"github.com/eolinker/goku-eosc/upstream"

	http_context "github.com/eolinker/goku-eosc/node/http-context"

	"github.com/eolinker/goku-eosc/service"
)

var (
	ErrorStructType = errors.New("error struct type")
)

type serviceWorker struct {
	id          string
	name        string
	driver      string
	desc        string
	timeout     time.Duration
	rewriteURL  string
	retry       int
	scheme      string
	proxyAddr   string
	proxyMethod string
	auths       []auth.IAuth
	upstream    upstream.IUpstream
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
	auths := make([]auth.IAuth, 0, len(data.Auth))
	for _, a := range data.Auth {
		if worker, has := workers[a]; has {
			ah, ok := worker.(auth.IAuth)
			if ok {
				auths = append(auths, ah)
			}
		}
	}
	s.desc = data.Desc
	s.timeout = time.Duration(data.Timeout) * time.Millisecond
	s.rewriteURL = data.RewriteURL
	s.retry = data.Retry
	s.scheme = data.Scheme
	s.proxyMethod = data.ProxyMethod
	s.auths = auths
	s.upstream = nil
	if worker, has := workers[data.Upstream]; has {
		u, ok := worker.(upstream.IUpstream)
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

//Name 返回服务名
func (s *serviceWorker) Name() string {
	return s.name
}

//Desc 返回服务的描述
func (s *serviceWorker) Desc() string {
	return s.desc
}

//Retry 返回服务的重试次数
func (s *serviceWorker) Retry() int {
	return s.retry
}

//Timeout 返回服务的超时时间
func (s *serviceWorker) Timeout() time.Duration {
	return s.timeout
}

//Scheme 返回服务的scheme
func (s *serviceWorker) Scheme() string {
	return s.scheme
}

//ProxyAddr 返回服务的代理地址
func (s *serviceWorker) ProxyAddr() string {
	return s.proxyAddr
}

func (s *serviceWorker) doAuth(ctx *http_context.Context) error {
	// 鉴权
	if len(s.auths) > 0 {
		validRequest := false
		for _, a := range s.auths {
			err := a.Auth(ctx)
			if err == nil {
				validRequest = true
				break
			}
			log.Error(err)
		}
		if !validRequest {
			return errors.New("invalid user")
		}
	}
	return nil
}

//Handle 将服务发送到负载
func (s *serviceWorker) Handle(ctx *http_context.Context, router service.IRouterEndpoint) error {
	// 构造context
	defer func() {
		if e := recover(); e != nil {
			log.Warn(e)
		}
		if ctx.GetStatus() == 0 {
			ctx.SetStatus(200)
		}
		ctx.Finish()
	}()
	err := s.doAuth(ctx)
	if err != nil {
		ctx.SetBody([]byte(err.Error()))
		ctx.SetStatus(403)
		return err
	}
	// 设置目标URL
	location, has := router.Location()
	path := s.rewriteURL
	if has && location.CheckType() == checker.CheckTypePrefix {
		path = recombinePath(string(ctx.RequestOrg().URI().Path()), location.Value(), s.rewriteURL)
	}
	if s.proxyMethod != "" {
		ctx.ProxyRequest().Header.SetMethod(s.proxyMethod)
	}
	body, err := ctx.BodyHandler.RawBody()
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
	ctx.SetResponse(response)
	return nil
}

func (s *serviceWorker) send(ctx *http_context.Context, serviceDetail service.IServiceDetail, uri string, query string) (*fasthttp.Response, error) {
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

//recombinePath 生成新的目标URL
func recombinePath(requestURL, location, targetURL string) string {
	newRequestURL := strings.Replace(requestURL, location, "", 1)
	return fmt.Sprintf("%s/%s", strings.TrimSuffix(targetURL, "/"), strings.TrimPrefix(newRequestURL, "/"))
}
