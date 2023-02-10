package manager

import (
	"fmt"
	"sync"
	"sync/atomic"

	"google.golang.org/grpc/peer"

	"github.com/eolinker/apinto/router"
	"google.golang.org/grpc"

	http_complete "github.com/eolinker/apinto/drivers/router/http-router/http-complete"
	eoscContext "github.com/eolinker/eosc/eocontext"
	"github.com/eolinker/eosc/log"
)

var _ IManger = (*Manager)(nil)
var notFound = new(NotFoundHandler)
var completeCaller = http_complete.NewHttpCompleteCaller()

type IManger interface {
	Set(id string, port int, hosts []string, service string, method string, append []AppendRule, router router.IRouterHandler) error
	Delete(id string)
}
type Manager struct {
	lock    sync.RWMutex
	matcher router.IMatcher

	routersData   IRouterData
	globalFilters atomic.Pointer[eoscContext.IChainPro]
}

func (m *Manager) SetGlobalFilters(globalFilters *eoscContext.IChainPro) {
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

func (m *Manager) FastHandler(port int, srv interface{}, stream grpc.ServerStream) {
	p, has := peer.FromContext(stream.Context())
	if !has {
		return
	}
	fmt.Println(p.Addr.String(), p.AuthInfo.AuthType())
	if m.matcher == nil {
		return
	}
}

type NotFoundHandler struct {
}

func (h *NotFoundHandler) Complete(ctx eoscContext.EoContext) error {
	panic("no implement")
}

func (h *NotFoundHandler) Finish(ctx eoscContext.EoContext) error {
	panic("no implement")
}
