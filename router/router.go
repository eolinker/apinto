package router

import (
	"github.com/eolinker/goku-eosc/router/checker"
)

type ISource interface {
	Get(cmd string)(string,bool)
}

type IRouter interface {
	Router(source ISource)(endpoint IEndpoint,has bool)
}

type Routers []IRouter

func (rs Routers) Router(source ISource) ( IEndpoint,  bool) {
	for _,r:=range rs{
		if target,has:=r.Router(source);has{
			return target,has
		}
	}
	return nil, false
}

type Node struct {
	cmd string

	equals map[string]IRouter

	checkers []checker.Checker
	nodes []IRouter

}



func (n *Node) Router(source ISource) ( IEndpoint,  bool) {

	v,has:=source.Get(n.cmd)

	if has{
		if child,ok:= n.equals[v];ok{
			if target,ok:=child.Router(source);ok{
				return target,true
			}
		}
	}

	for i,c:=range n.checkers{
		if c.Check(v,has){
			if target,ok:=n.nodes[i].Router(source);ok{
				return target,true
			}
		}
	}

	return nil,false

}

type NodeShut struct {
	next IRouter
	endpoint IEndpoint
}

func (n *NodeShut) Router(source ISource) (  IEndpoint,   bool) {
	if e ,has:=n.next.Router(source);has{
		return e,has
	}
	return n.endpoint,true
}

func NewNodeShut(next IRouter, endpoint IEndpoint) IRouter {
	return &NodeShut{next: next, endpoint: endpoint}
}