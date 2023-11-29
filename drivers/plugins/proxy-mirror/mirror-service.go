package proxy_mirror

import (
	"errors"
	"strings"
	"time"

	"github.com/eolinker/eosc/eocontext"

	"github.com/eolinker/apinto/discovery"
)

var (
	errNoValidNode                          = errors.New("no valid node")
	_              eocontext.BalanceHandler = (*mirrorService)(nil)
)

type mirrorService struct {
	app      discovery.IApp
	scheme   string
	passHost eocontext.PassHostMod
	host     string
	timeout  time.Duration
}

func (m *mirrorService) Select(ctx eocontext.EoContext) (eocontext.INode, int, error) {
	for i, node := range m.app.Nodes() {
		if node.Status() != eocontext.Down {
			return node, i, nil
		}
	}
	return nil, 0, errNoValidNode
}

func (m *mirrorService) stop() {
	m.app.Close()
}

func newMirrorService(target, passHost, host string, timeout time.Duration) *mirrorService {
	idx := strings.Index(target, "://")
	scheme := target[:idx]
	addr := target[idx+3:]

	idx = strings.Index(addr, ":")
	app, _ := defaultProxyDiscovery.GetApp(addr)

	var mode eocontext.PassHostMod
	switch passHost {
	case modePass:
		mode = eocontext.PassHost
	case modeNode:
		mode = eocontext.NodeHost
	case modeRewrite:
		mode = eocontext.ReWriteHost
	}

	return &mirrorService{
		app:      app,
		scheme:   scheme,
		passHost: mode,
		host:     host,
		timeout:  timeout,
	}
}

func (m *mirrorService) Nodes() []eocontext.INode {
	return m.app.Nodes()
}

func (m *mirrorService) Scheme() string {
	return m.scheme
}

func (m *mirrorService) TimeOut() time.Duration {
	return m.timeout
}

func (m *mirrorService) PassHost() (eocontext.PassHostMod, string) {
	return m.passHost, m.host
}
