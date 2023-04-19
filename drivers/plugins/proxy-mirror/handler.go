package proxy_mirror

import (
	"github.com/eolinker/eosc/eocontext"
	"github.com/eolinker/eosc/log"
)

type proxyMirrorCompleteHandler struct {
	orgComplete eocontext.CompleteHandler
	service     *mirrorService
}

func newMirrorHandler(eoCtx eocontext.EoContext, service *mirrorService) (eocontext.CompleteHandler, error) {
	handler := &proxyMirrorCompleteHandler{
		orgComplete: eoCtx.GetComplete(),
		service:     service,
	}

	return handler, nil
}

func (p *proxyMirrorCompleteHandler) Complete(ctx eocontext.EoContext) error {

	//先执行原始Complete, 再执行镜像请求的Complete
	orgErr := p.orgComplete.Complete(ctx)
	cloneCtx, err := ctx.Clone()
	if err != nil {
		log.Warn(err)
		return orgErr
	}

	cloneCtx.SetBalance(p.service)
	cloneCtx.SetUpstreamHostHandler(p.service)

	go func() {
		err = p.orgComplete.Complete(cloneCtx)
		if err != nil {
			log.Error(err)
		}
	}()

	return orgErr
}
