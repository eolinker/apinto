package discovery

import (
	"fmt"
	"github.com/eolinker/eosc/eocontext"
)

// NodeStatus 节点状态类型
type NodeStatus = eocontext.NodeStatus

const (
	//Running 节点运行中状态
	Running = eocontext.Running
	//Down 节点不可用状态
	Down = eocontext.Down
	//Leave 节点离开状态
	Leave = eocontext.Leave
)

type INode interface {
	IP() string
	ID() string
	Addr() string
	Port() int
	Status() NodeStatus
	Up()
	Down()
	Leave()
}
type _INodeStatusCheck interface {
	status(status NodeStatus) NodeStatus
}
type _BaseNode struct {
	id   string
	ip   string
	port int

	status        NodeStatus
	statusChecker _INodeStatusCheck
}

func newBaseNode(id string, ip string, port int, statusChecker _INodeStatusCheck) *_BaseNode {
	return &_BaseNode{id: id, ip: ip, port: port, status: Running, statusChecker: statusChecker}
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

	return n.statusChecker.status(n.status)
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
	n.status = Running
}

// Down 将节点状态置为不可用
func (n *_BaseNode) Down() {
	n.status = Down
}

// Leave 将节点状态置为离开
func (n *_BaseNode) Leave() {
	n.status = Leave
}

// Attrs 属性集合
type Attrs = eocontext.Attrs
type Node struct {
	INode
	label Attrs
}

func NewNode(INode INode, label Attrs) eocontext.INode {
	return &Node{INode: INode, label: label}
}

func (n *Node) GetAttrs() eocontext.Attrs {
	return n.label
}

func (n *Node) GetAttrByName(name string) (string, bool) {
	v, h := n.label[name]
	return v, h
}
