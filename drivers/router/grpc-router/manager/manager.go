package manager

import (
	"sync"
	"sync/atomic"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	grpc_context "github.com/eolinker/apinto/node/grpc-context"

	"github.com/eolinker/apinto/router"
	"google.golang.org/grpc"

	"github.com/eolinker/eosc/eocontext"
	"github.com/eolinker/eosc/log"
)

var _ IManger = (*Manager)(nil)
var completeCaller = NewCompleteCaller()

type IManger interface {
	Set(id string, port int, hosts []string, service string, method string, append []AppendRule, router router.IRouterHandler) error
	Delete(id string)
}
type Manager struct {
	lock    sync.RWMutex
	matcher router.IMatcher

	routersData   IRouterData
	globalFilters atomic.Pointer[eocontext.IChainPro]
}

func (m *Manager) SetGlobalFilters(globalFilters *eocontext.IChainPro) {
	m.globalFilters.Store(globalFilters)
}

// NewManager 创建路由管理器
func NewManager() *Manager {
	return &Manager{routersData: new(RouterData)}
}

func (m *Manager) Set(id string, port int, hosts []string, service string, method string, append []AppendRule, router router.IRouterHandler) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	routersData := m.routersData.Set(id, port, hosts, service, method, append, router)
	matchers, err := routersData.Parse()
	if err != nil {
		log.Error("parse router data error: ", err)
		return err
	}
	m.matcher = matchers
	m.routersData = routersData
	return nil
}

func (m *Manager) Delete(id string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	routersData := m.routersData.Delete(id)
	matchers, err := routersData.Parse()
	if err != nil {
		log.Errorf("delete router:%s %s", id, err.Error())
		return
	}
	m.matcher = matchers
	m.routersData = routersData
	return
}

func (m *Manager) FastHandler(port int, srv interface{}, stream grpc.ServerStream) error {
	ctx := grpc_context.NewContext(srv, stream)
	if m.matcher == nil {
		return status.Error(codes.NotFound, "not found")
	}

	r, has := m.matcher.Match(port, ctx.Request())
	if !has {
		errHandler := NewErrHandler(status.Error(codes.NotFound, "not found"))
		ctx.SetFinish(errHandler)
		ctx.SetCompleteHandler(errHandler)
		globalFilters := m.globalFilters.Load()
		if globalFilters != nil {
			(*globalFilters).Chain(ctx, completeCaller)
		}
	} else {
		log.Debug("match has:", port)
		r.Serve(ctx)
	}

	finishHandler := ctx.GetFinish()
	if finishHandler != nil {
		return finishHandler.Finish(ctx)
	}
	return nil
}

type ErrHandler struct {
	err error
}

func (e *ErrHandler) Complete(ctx eocontext.EoContext) error {
	return e.err
}

func NewErrHandler(err error) *ErrHandler {
	return &ErrHandler{err: err}
}

func (e *ErrHandler) Finish(ctx eocontext.EoContext) error {
	return e.err
}
