package proxy_mirror

import (
	"github.com/eolinker/eosc/eocontext"
	dubbo2_context "github.com/eolinker/eosc/eocontext/dubbo2-context"
	grpc_context "github.com/eolinker/eosc/eocontext/grpc-context"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	log "github.com/eolinker/goku-api-gateway/goku-log"
)

type proxyMirrorCompleteHandler struct {
	orgComplete    eocontext.CompleteHandler
	mirrorComplete eocontext.CompleteHandler
}

func newMirrorHandler(eoCtx eocontext.EoContext, proxyCfg *Config) (eocontext.CompleteHandler, error) {
	handler := &proxyMirrorCompleteHandler{
		orgComplete: eoCtx.GetComplete(),
	}

	if _, success := eoCtx.(http_service.IHttpContext); success {
		handler.mirrorComplete = newHttpMirrorComplete(proxyCfg)
	} else if _, success = eoCtx.(grpc_context.IGrpcContext); success {
		handler.mirrorComplete = newGrpcMirrorComplete(proxyCfg)
	} else if _, success = eoCtx.(dubbo2_context.IDubbo2Context); success {
		handler.mirrorComplete = newDubbo2MirrorComplete(proxyCfg)
	} else if _, success = eoCtx.(http_service.IWebsocketContext); success {
		handler.mirrorComplete = newWebsocketMirrorComplete(proxyCfg)
	} else {
		return nil, ErrUnsupportedType
	}

	return handler, nil
}

func (p *proxyMirrorCompleteHandler) Complete(ctx eocontext.EoContext) error {
	cloneCtx, err := ctx.Clone()
	if err != nil {
		return err
	}

	//先执行原始Complete, 再执行镜像请求的Complete
	err = p.orgComplete.Complete(ctx)

	go func() {
		mErr := p.mirrorComplete.Complete(cloneCtx)
		if mErr != nil {
			log.Error(mErr)
		}
	}()

	return err
}
