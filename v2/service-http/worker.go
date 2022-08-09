package service_http

import (
	"errors"
	"fmt"
	"github.com/eolinker/apinto/service"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
	"time"
)

var (
	ErrorNeedUpstream = errors.New("need upstream")

	ErrorInvalidDiscovery = errors.New("invalid Discovery")
)

type serviceWorker struct {
	Service
	id     string
	name   string
	driver string
}

func (s *Service) Complete(org eocontext.EoContext) error {

	ctx, err := http_service.Assert(org)
	if err != nil {
		return err
	}
	//设置响应开始时间
	proxyTime := time.Now()

	defer func() {
		//设置原始响应状态码
		ctx.Response().SetProxyStatus(ctx.Response().StatusCode(), "")
		//设置上游响应时间, 单位为毫秒
		ctx.WithValue("response_time", time.Now().Sub(proxyTime).Milliseconds())
	}()

	var lastErr error
	for doTrice := s.retry + 1; doTrice > 0; doTrice-- {

		node, err := s.balanceHandler.Next()
		if err != nil {
			return err
		}

		scheme := s.scheme

		log.Debug("node: ", node.Addr())
		addr := fmt.Sprintf("%s://%s", scheme, node.Addr())
		lastErr = ctx.SendTo(addr, s.timeout)
		if lastErr == nil {
			return nil
		}
		log.Error("http upstream send error: ", lastErr)
	}

	return lastErr
}

func (s *serviceWorker) Finish(org eocontext.EoContext) error {
	ctx, err := http_service.Assert(org)
	if err != nil {
		return err
	}
	ctx.FastFinish()
	return nil
}

//Id 返回服务实例 worker id
func (s *serviceWorker) Id() string {
	return s.id
}

func (s *serviceWorker) Start() error {
	return nil
}

func (s *serviceWorker) Stop() error {

	return nil
}

//CheckSkill 检查目标能力是否存在
func (s *serviceWorker) CheckSkill(skill string) bool {
	return service.CheckSkill(skill)
}
