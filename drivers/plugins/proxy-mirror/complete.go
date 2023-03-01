package proxy_mirror

import (
	"fmt"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
	"time"
)

type httpComplete struct {
	proxyCfg *Config
}

func newHttpMirrorComplete(proxyCfg *Config) eocontext.CompleteHandler {
	return &httpComplete{
		proxyCfg: proxyCfg,
	}
}

func (h *httpComplete) Complete(org eocontext.EoContext) error {
	ctx, err := http_service.Assert(org)
	if err != nil {
		return err
	}
	//设置响应开始时间
	proxyTime := time.Now()

	defer func() {
		//设置原始响应状态码
		ctx.Response().SetProxyStatus(ctx.Response().StatusCode(), "")
		//设置上游响应总时间, 单位为毫秒
		//ctx.WithValue("response_time", time.Now().Sub(proxyTime).Milliseconds())
		ctx.Response().SetResponseTime(time.Now().Sub(proxyTime))
		ctx.SetLabel("handler", "proxy")
	}()

	var lastErr error

	timeOut := time.Duration(h.proxyCfg.Timeout) * time.Millisecond

	//构造addr
	path := ctx.Request().URI().Path()
	if h.proxyCfg.Path != "" {
		switch h.proxyCfg.PathMode {
		case pathModeReplace:
			path = h.proxyCfg.Path
		case pathModePrefix:
			path = fmt.Sprintf("%s%s", h.proxyCfg.Path, path)
		}
	}
	addr := fmt.Sprintf("%s%s", h.proxyCfg.Host, path)

	lastErr = ctx.SendTo(addr, timeOut)
	if lastErr == nil {
		return nil
	}
	log.Error("http proxyMirror send error: ", lastErr)

	return lastErr
}

type dubbo2Complete struct {
	proxyCfg *Config
}

func newDubbo2MirrorComplete(proxyCfg *Config) eocontext.CompleteHandler {
	return &httpComplete{
		proxyCfg: proxyCfg,
	}
}

func (d *dubbo2Complete) Complete(ctx eocontext.EoContext) error {
	//TODO implement me
	return nil
}

type grpcComplete struct {
	proxyCfg *Config
}

func newGrpcMirrorComplete(proxyCfg *Config) eocontext.CompleteHandler {
	return &httpComplete{
		proxyCfg: proxyCfg,
	}
}

func (g *grpcComplete) Complete(ctx eocontext.EoContext) error {
	//TODO implement me
	return nil
}

type websocketComplete struct {
	proxyCfg *Config
}

func newWebsocketMirrorComplete(proxyCfg *Config) eocontext.CompleteHandler {
	return &httpComplete{
		proxyCfg: proxyCfg,
	}
}

func (w *websocketComplete) Complete(ctx eocontext.EoContext) error {
	//TODO implement me
	return nil
}
