package discovery

import (
	"fmt"
	"sync"
)

type INodes interface {
	Get(scheme, ip string, port int) INode
	All() []INode
	remove(id ...string)
}

type nodes struct {
	locker sync.RWMutex
	m      map[string]INode
}

func (n *nodes) remove(ids ...string) {
	n.locker.Lock()
	defer n.locker.Unlock()

	for _, id := range ids {
		delete(n.m, id)
	}
}

func (n *nodes) Get(ip string, port int) INode {
	id := fmt.Sprintf("%s:%d", ip, port)
	n.locker.RLock()
	node, has := n.m[id]
	n.locker.RUnlock()
	if has {
		return node
	}
	n.locker.Lock()
	defer n.locker.Unlock()
	node, has = n.m[id]
	if has {
		return node
	}

	n.m[id] = newBaseNode(ip, port)
	return n.m[id]
}

func (n *nodes) All() []INode {

	n.locker.RLock()
	ls := make([]INode, 0, len(n.m))
	for _, node := range n.m {
		ls = append(ls, node)
	}
	n.locker.RUnlock()
	return ls
}
