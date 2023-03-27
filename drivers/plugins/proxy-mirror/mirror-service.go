package proxy_mirror

import (
	"errors"
	"github.com/eolinker/apinto/discovery"
	"github.com/eolinker/eosc/eocontext"
	"strings"
	"time"
)

var (
	errNoValidNode = errors.New("no valid node")
)

type mirrorService struct {
	app      discovery.IApp
	scheme   string
	passHost eocontext.PassHostMod
	host     string
	timeout  time.Duration
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

