package manager

import (
	"dubbo.apache.org/dubbo-go/v3/protocol"
	"dubbo.apache.org/dubbo-go/v3/protocol/invocation"
	"errors"
	dubbo2_context "github.com/eolinker/apinto/node/dubbo2-context"
	"github.com/eolinker/apinto/router"
	eoscContext "github.com/eolinker/eosc/eocontext"
	"github.com/eolinker/eosc/log"
	"sync"
	"sync/atomic"
)

var _ IManger = (*dubboManger)(nil)

var completeCaller = NewCompleteCaller()

type IManger interface {
	Set(id string, port int, serviceName, methodName string, rule []AppendRule, handler router.IRouterHandler) error
	Delete(id string)
}

func (d *dubboManger) SetGlobalFilters(globalFilters *eoscContext.IChainPro) {
	d.globalFilters.Store(globalFilters)
}

func NewManager() *dubboManger {
	return &dubboManger{
		matcher:       nil,
		routersData:   new(RouterData),
		globalFilters: atomic.Pointer[eoscContext.IChainPro]{},
	}
}

type dubboManger struct {
	lock          sync.RWMutex
	matcher       router.IMatcher
	routersData   IRouterData
	globalFilters atomic.Pointer[eoscContext.IChainPro]
}

func (d *dubboManger) Set(id string, port int, serviceName, methodName string, rule []AppendRule, handler router.IRouterHandler) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	routersData := d.routersData.Set(id, port, serviceName, methodName, rule, handler)
	matchers, err := routersData.Parse()
	if err != nil {
		log.Error("parse router data error: ", err)
		return err
	}
	d.matcher = matchers
	d.routersData = routersData
	return nil
}

func (d *dubboManger) Delete(id string) {
	d.lock.Lock()
	defer d.lock.Unlock()
	routersData := d.routersData.Delete(id)
	matchers, err := routersData.Parse()
	if err != nil {
		log.Errorf("delete router:%s %s", id, err.Error())
		return
	}

	d.matcher = matchers
	d.routersData = routersData
	return
}

func (d *dubboManger) Handler(port int, req *invocation.RPCInvocation) protocol.RPCResult {

	ctx := dubbo2_context.NewContext(req, port)

	match, has := d.matcher.Match(port, ctx.HeaderReader())
	if !has {
		errHandler := NewErrHandler(errors.New("not found"))
		ctx.SetFinish(errHandler)
		ctx.SetCompleteHandler(errHandler)

		globalFilters := d.globalFilters.Load()
		if globalFilters != nil {
			if err := (*globalFilters).Chain(ctx, completeCaller); err != nil {
				ctx.Response().SetBody(Dubbo2ErrorResult(err))
			}
		}

	} else {
		log.Debug("match has:", port)
		match.ServeHTTP(ctx)
	}

	finish := ctx.GetFinish()
	err := finish.Finish(ctx)
	if err != nil {
		ctx.Response().SetBody(Dubbo2ErrorResult(err))
	}

	rpcResult, ok := ctx.Response().GetBody().(protocol.RPCResult)
	if !ok {
		rpcResult = Dubbo2ErrorResult(errors.New("no result"))
	}

	return rpcResult
}

type ErrHandler struct {
	err error
}

func (e *ErrHandler) Complete(ctx eoscContext.EoContext) error {
	return e.err
}

func NewErrHandler(err error) *ErrHandler {
	return &ErrHandler{err: err}
}

func (e *ErrHandler) Finish(ctx eoscContext.EoContext) error {
	return e.err
}
