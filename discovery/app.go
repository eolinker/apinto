package discovery

import (
	"sync"
	"sync/atomic"

	"github.com/eolinker/eosc/eocontext"
)

var (
	_ IAppAgent = (*_AppAgent)(nil)
	_ IApp      = (*_App)(nil)
)

type IAppAgent interface {
	reset(nodes []eocontext.INode)
	Agent() IApp
}

type IApp interface {
	Nodes() []eocontext.INode
	Close()
}

type _AppAgent struct {
	locker sync.RWMutex
	nodes  []eocontext.INode
	use    int64
}

func (a *_AppAgent) Agent() IApp {
	atomic.AddInt64(&a.use, 1)
	return &_App{_AppAgent: a, isClose: 0}
}

type _App struct {
	*_AppAgent

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
