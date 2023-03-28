package discovery

import (
	"fmt"
)

var (
	_ INodes = (*appContainer)(nil)
)

type INodes interface {
	Get(ip string, port int) INode
	All() []INode
	SetHealthCheck(isHealthCheck bool)
}

func (ac *appContainer) Get(ip string, port int) INode {
	id := fmt.Sprintf("%s:%d", ip, port)

	node, has := ac.nodes.Get(id)

	if has {
		return node
	}
	ac.lock.Lock()
	defer ac.lock.Unlock()
	node, has = ac.nodes.Get(id)
	if has {
		return node
	}

	ac.nodes.Set(id, newBaseNode(ip, port, ac))
	node, _ = ac.nodes.Get(id)
	return node
}

func (ac *appContainer) All() []INode {

	return ac.nodes.List()
}
