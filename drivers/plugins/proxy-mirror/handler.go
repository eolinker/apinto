package proxy_mirror

import (
	"github.com/eolinker/eosc/eocontext"
	"github.com/eolinker/eosc/log"
)

type proxyMirrorCompleteHandler struct {
	orgComplete eocontext.CompleteHandler
	app         eocontext.EoApp
}

func newMirrorHandler(eoCtx eocontext.EoContext, app eocontext.EoApp) (eocontext.CompleteHandler, error) {
	handler := &proxyMirrorCompleteHandler{
		orgComplete: eoCtx.GetComplete(),
		app:         app,
	}

	return handler, nil
}

func (p *proxyMirrorCompleteHandler) Complete(ctx eocontext.EoContext) error {
	cloneCtx, err := ctx.Clone()

	//先执行原始Complete, 再执行镜像请求的Complete
	orgErr := p.orgComplete.Complete(ctx)

	if err == nil {
		cloneCtx.SetApp(p.app)

		go func() {
			err = p.orgComplete.Complete(cloneCtx)
			if err != nil {
				log.Error(err)
			}
		}()
	} else {
		log.Error(err)
	}

	return orgErr
}
