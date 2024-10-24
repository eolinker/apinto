package ai_provider

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/eolinker/eosc/eocontext"
)

var _ eocontext.INode = (*_BaseNode)(nil)

func NewBaseNode(id string, ip string, port int) *_BaseNode {
	return &_BaseNode{id: id, ip: ip, port: port}
}

type _BaseNode struct {
	id     string
	ip     string
	port   int
	status eocontext.NodeStatus
}

func (n *_BaseNode) GetAttrs() eocontext.Attrs {
	return map[string]string{}
}

func (n *_BaseNode) GetAttrByName(name string) (string, bool) {
	return "", false
}

func (n *_BaseNode) ID() string {
	return n.id
}

func (n *_BaseNode) IP() string {
	return n.ip
}

func (n *_BaseNode) Port() int {
	return n.port
}

func (n *_BaseNode) Status() eocontext.NodeStatus {

	return n.status
}

// Addr 返回节点地址
func (n *_BaseNode) Addr() string {
	if n.port == 0 {
		return n.ip
	}
	return fmt.Sprintf("%s:%d", n.ip, n.port)
}

// Up 将节点状态置为运行中
func (n *_BaseNode) Up() {
	n.status = eocontext.Running
}

// Down 将节点状态置为不可用
func (n *_BaseNode) Down() {
	n.status = eocontext.Down
}

// Leave 将节点状态置为离开
func (n *_BaseNode) Leave() {
	n.status = eocontext.Leave
}

func NewBalanceHandler(id string, base string, timeout time.Duration) (eocontext.BalanceHandler, error) {
	u, err := url.Parse(base)
	if err != nil {
		return nil, err
	}
	hosts := strings.Split(u.Host, ":")
	ip := hosts[0]
	port := 80
	if u.Scheme == "https" {
		port = 443
	}
	if len(hosts) > 1 {
		port, _ = strconv.Atoi(hosts[1])
	}

	return &_BalanceHandler{scheme: u.Scheme, timeout: timeout, nodes: []eocontext.INode{NewBaseNode(id, ip, port)}}, nil
}

type _BalanceHandler struct {
	scheme  string
	timeout time.Duration
	nodes   []eocontext.INode
}

func (b *_BalanceHandler) Select(ctx eocontext.EoContext) (eocontext.INode, int, error) {
	if len(b.nodes) == 0 {
		return nil, 0, nil
	}
	return b.nodes[0], 0, nil
}

func (b *_BalanceHandler) Scheme() string {
	return b.scheme
}

func (b *_BalanceHandler) TimeOut() time.Duration {
	return b.timeout
}

func (b *_BalanceHandler) Nodes() []eocontext.INode {
	return b.nodes
}
