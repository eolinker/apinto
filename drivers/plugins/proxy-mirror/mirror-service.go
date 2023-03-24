package proxy_mirror

import (
	"errors"
	"fmt"
	"github.com/eolinker/apinto/discovery"
	"github.com/eolinker/eosc/eocontext"
	"strconv"
	"strings"
	"time"
)

var (
	errNoValidNode = errors.New("no valid node")
)

type mirrorService struct {
	scheme   string
	passHost eocontext.PassHostMod
	host     string
	timeout  time.Duration
	nodes    []eocontext.INode
}

func newMirrorService(target, passHost, host string, timeout time.Duration) *mirrorService {
	labels := map[string]string{}

	idx := strings.Index(target, "://")
	scheme := target[:idx]
	addr := target[idx+3:]

	idx = strings.Index(addr, ":")
	ip := addr
	port := 0
	if idx > 0 {
		ip = addr[:idx]
		portStr := addr[idx+1:]
		port, _ = strconv.Atoi(portStr)
	}

	inode := discovery.n(labels, fmt.Sprintf("%s:%d", ip, port), ip, port)

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
		scheme:   scheme,
		passHost: mode,
		host:     host,
		timeout:  timeout,
		nodes:    []eocontext.INode{inode},
	}
}

func (m *mirrorService) Nodes() []eocontext.INode {
	return m.nodes
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

func (m *mirrorService) Select(ctx eocontext.EoContext) (eocontext.INode, error) {
	if len(m.nodes) < 1 {
		return nil, errNoValidNode
	}
	return m.nodes[0], nil
}
