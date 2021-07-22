package discovery

import (
	"fmt"
)

//NewNode 创建新节点
func NewNode(labels map[string]string, id string, ip string, port int) INode {
	return &node{labels: labels, id: id, ip: ip, port: port, status: Running}
}

type node struct {
	labels Attrs
	id     string
	ip     string
	port   int
	status NodeStatus
}

//GetAttrs 获取节点属性集合
func (n *node) GetAttrs() Attrs {
	return n.labels
}

//GetAttrByName 通过属性名获取节点属性
func (n *node) GetAttrByName(name string) (string, bool) {
	v, ok := n.labels[name]
	return v, ok
}

//IP 返回节点IP
func (n *node) IP() string {
	return n.ip
}

//Port 返回节点端口
func (n *node) Port() int {
	return n.port
}

//ID 返回节点ID
func (n *node) ID() string {
	return n.id
}

//Status 返回节点状态
func (n *node) Status() NodeStatus {
	return n.status
}

//Labels 返回节点标签集合
func (n *node) Labels() map[string]string {
	return n.labels
}

//Addr 返回节点地址
func (n *node) Addr() string {
	if n.port == 0 {
		return n.ip
	}
	return fmt.Sprintf("%s:%d", n.ip, n.port)
}

//Up 将节点状态置为运行中
func (n *node) Up() {
	n.status = Running
}

//Down 将节点状态置为不可用
func (n *node) Down() {
	n.status = Down
}

//Leave 将节点状态置为离开
func (n *node) Leave() {
	n.status = Leave
}
