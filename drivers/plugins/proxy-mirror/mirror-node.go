package proxy_mirror

import (
	"errors"
	"fmt"
	"github.com/eolinker/eosc/eocontext"
)

var (
	errNoValidNode = errors.New("no valid node")
)

type node struct {
	labels eocontext.Attrs
	id     string
	ip     string
	port   int
	status eocontext.NodeStatus
}

// newNode 创建新节点
func newNode(labels map[string]string, id string, ip string, port int) eocontext.INode {
	return &node{labels: labels, id: id, ip: ip, port: port, status: eocontext.Running}
}

// GetAttrs 获取节点属性集合
func (n *node) GetAttrs() eocontext.Attrs {
	return n.labels
}

// GetAttrByName 通过属性名获取节点属性
func (n *node) GetAttrByName(name string) (string, bool) {
	v, ok := n.labels[name]
	return v, ok
}

// IP 返回节点IP
func (n *node) IP() string {
	return n.ip
}

// Port 返回节点端口
func (n *node) Port() int {
	return n.port
}

// ID 返回节点ID
func (n *node) ID() string {
	return n.id
}

// Status 返回节点状态
func (n *node) Status() eocontext.NodeStatus {
	return n.status
}

// Labels 返回节点标签集合
func (n *node) Labels() map[string]string {
	return n.labels
}

// Addr 返回节点地址
func (n *node) Addr() string {
	if n.port == 0 {
		return n.ip
	}
	return fmt.Sprintf("%s:%d", n.ip, n.port)
}

// Up 将节点状态置为运行中
func (n *node) Up() {
	n.status = eocontext.Running
}

// Down 将节点状态置为不可用
func (n *node) Down() {
	n.status = eocontext.Down
}

// Leave 将节点状态置为离开
func (n *node) Leave() {
	n.status = eocontext.Leave
}
