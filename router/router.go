package router

import (
	"github.com/eolinker/goku/checker"
)

//ISource 实现了从源对象中获取相应指标的值的方法
type ISource interface {
	Get(cmd string) (string, bool)
}

//IRouter 实现了Router方法
type IRouter interface {
	Router(source ISource) (endpoint IEndPoint, has bool)
}

//Routers 路由树结构体
type Routers []IRouter

//Router 路由获得端点
func (rs Routers) Router(source ISource) (IEndPoint, bool) {
	for _, r := range rs {
		if target, has := r.Router(source); has {
			return target, has
		}
	}
	return nil, false
}

//Node 路由树中指标节点结构体
type Node struct {
	cmd string

	equals map[string]IRouter //存放使用全等匹配的指标节点

	checkers []checker.Checker //按优先顺序存放除全等匹配外的checker，顺序与nodes对应
	nodes    []IRouter         //按优先顺序存放使用除全等匹配外的指标节点
}

//Router 路由方法
func (n *Node) Router(source ISource) (IEndPoint, bool) {

	v, has := source.Get(n.cmd)

	if has {
		if child, ok := n.equals[v]; ok {
			if target, ok := child.Router(source); ok {
				return target, true
			}
		}
	}

	for i, c := range n.checkers {
		if c.Check(v, has) {
			if target, ok := n.nodes[i].Router(source); ok {
				return target, true
			}
		}
	}

	return nil, false

}

//NodeShut NodeShut
type NodeShut struct {
	next     IRouter
	endpoint IEndPoint
}

//Router 路由方法
func (n *NodeShut) Router(source ISource) (IEndPoint, bool) {
	if e, has := n.next.Router(source); has {
		return e, has
	}
	return n.endpoint, true
}

//NewNodeShut 创建NewNodeShut，暂未用到
func NewNodeShut(next IRouter, endpoint IEndPoint) IRouter {
	return &NodeShut{next: next, endpoint: endpoint}
}
