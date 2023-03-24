package discovery

import (
	"github.com/eolinker/eosc/eocontext"
	"sync"
	"sync/atomic"
	"time"
)

var (
	_ IAppAgent = (*_AppAgent)(nil)
	_ IApp      = (*_App)(nil)
)

type IAppAgent interface {
	reset(nodes []eocontext.INode)
	Agent(scheme string, timeout time.Duration) IApp
}

type IApp interface {
	eocontext.EoApp
	Close()
}

type _AppAgent struct {
	id      string
	locker  sync.RWMutex
	nodes   []eocontext.INode
	timeout time.Duration
	use     int64
}

func (a *_AppAgent) Agent(scheme string, timeout time.Duration) IApp {
	atomic.AddInt64(&a.use, 1)
	return &_App{_AppAgent: a, scheme: scheme, timeout: timeout, isClose: 0}
}

type _App struct {
	*_AppAgent
	scheme  string
	timeout time.Duration
	isClose int32
}

func (a *_App) Close() {
	if atomic.SwapInt32(&a.isClose, 1) == 0 {
		atomic.AddInt64(&a.use, -1)
	}

}

func newApp(nodes []eocontext.INode) *_AppAgent {

	return &_AppAgent{nodes: nodes}
}

func (a *_AppAgent) reset(nodes []eocontext.INode) {

	a.locker.Lock()
	defer a.locker.Unlock()
	a.nodes = nodes
}

func (a *_AppAgent) Nodes() []eocontext.INode {
	a.locker.RLock()
	defer a.locker.RUnlock()
	l := make([]eocontext.INode, len(a.nodes))
	copy(l, a.nodes)
	return l
}

func (a *_App) Scheme() string {
	return a.scheme
}

func (a *_App) TimeOut() time.Duration {
	return a.timeout
}
