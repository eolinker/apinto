package discovery

import (
	"fmt"
)

func NewNode(labels map[string]string, id string, ip string, port int) *Node {
	return &Node{labels: labels, id: id, ip: ip, port: port, status: Running}
}

type Node struct {
	labels Attrs
	id     string
	ip     string
	port   int
	status NodeStatus
}

func (n *Node) GetAttrs() Attrs {
	return n.labels
}

func (n *Node) GetAttrByName(name string) (string, bool) {
	v, ok := n.labels[name]
	return v, ok
}

func (n *Node) Ip() string {
	return n.ip
}

func (n *Node) Port() int {
	return n.port
}

func (n *Node) Id() string {
	return n.id
}

func (n *Node) Status() NodeStatus {
	return n.status
}

func (n *Node) Labels() map[string]string {
	return n.labels
}

func (n *Node) Addr() string {
	if n.port == 0 {
		return n.ip
	}
	return fmt.Sprintf("%s:%d", n.ip, n.port)
}

func (n *Node) Up() {
	n.status = Running
}

func (n *Node) Down() {
	n.status = Down
}

func (n *Node) Leave() {
	n.status = Leave
}
