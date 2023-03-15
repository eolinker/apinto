package dubbo2_router

import (
	"errors"
	"time"

	"github.com/eolinker/apinto/drivers/router/dubbo2-router/manager"
	"github.com/eolinker/apinto/router"
	"github.com/eolinker/apinto/service"
	"github.com/eolinker/eosc/eocontext"
	dubbo2_context "github.com/eolinker/eosc/eocontext/dubbo2-context"
)

var _ router.IRouterHandler = (*dubboHandler)(nil)

type dubboHandler struct {
	completeHandler eocontext.CompleteHandler
	finishHandler   eocontext.FinishHandler
	routerName      string
	routerId        string
	serviceName     string
	disable         bool
	service         service.IService
	filters         eocontext.IChainPro
	retry           int
	timeout         time.Duration
}

var completeCaller = manager.NewCompleteCaller()

func (d *dubboHandler) ServeHTTP(ctx eocontext.EoContext) {

	dubboCtx, err := dubbo2_context.Assert(ctx)
	if err != nil {
		return
	}

	if d.disable {
		dubboCtx.Response().SetBody(manager.Dubbo2ErrorResult(errors.New("router disable")))
		return
	}

	//set retry timeout
	ctx.WithValue(eocontext.CtxKeyRetry, d.retry)
	ctx.WithValue(eocontext.CtxKeyTimeout, d.timeout)

	//Set Label
	ctx.SetLabel("api", d.routerName)
	ctx.SetLabel("api_id", d.routerId)
	ctx.SetLabel("service", d.serviceName)
	ctx.SetLabel("service_id", d.service.Id())
	ctx.SetLabel("ip", dubboCtx.HeaderReader().RemoteIP())

	ctx.SetCompleteHandler(d.completeHandler)
	ctx.SetFinish(d.finishHandler)
	ctx.SetApp(d.service)
	ctx.SetBalance(d.service)
	ctx.SetUpstreamHostHandler(d.service)

	_ = d.filters.Chain(ctx, completeCaller)

}
