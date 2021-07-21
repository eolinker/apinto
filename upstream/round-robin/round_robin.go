package round_robin

import (
	"errors"
	"strconv"

	"github.com/eolinker/goku-eosc/upstream/balance"

	"github.com/eolinker/goku-eosc/discovery"
)

const (
	name = "round-robin"
)

func Register() {
	balance.Register(name, newRoundRobinFactory())
}

func newRoundRobinFactory() *roundRobinFactory {
	return &roundRobinFactory{}
}

type roundRobinFactory struct {
}

func (r *roundRobinFactory) Create(app discovery.IApp) (balance.IBalanceHandler, error) {
	rr := newRoundRobin(app.Nodes())
	return rr, nil
}

type node struct {
	weight int
	discovery.INode
}

type roundRobin struct {
	// nodes 节点列表
	nodes []node
	// 节点数量
	size int
	// index 当前索引
	index int
	// gcdWeight 权重最大公约数
	gcdWeight int
	// maxWeight 权重最大值
	maxWeight int
	cw        int
}

func (r *roundRobin) Next() (discovery.INode, error) {
	for {
		r.index = (r.index + 1) % r.size
		if r.index == 0 {
			r.cw = r.cw - r.gcdWeight
			if r.cw <= 0 {
				r.cw = r.maxWeight
				if r.cw == 0 {
					return nil, errors.New("")
				}
			}
		}
		if r.nodes[r.index].weight >= r.cw {
			if r.nodes[r.index].Status() == discovery.Down {
				continue
			}
			return r.nodes[r.index], nil
		}
	}
}

func newRoundRobin(nodes []discovery.INode) *roundRobin {
	size := len(nodes)
	r := &roundRobin{
		nodes: make([]node, 0, size),
		size:  size,
	}
	for i, n := range nodes {

		weight, _ := n.GetAttrByName("weight")
		w, _ := strconv.Atoi(weight)
		if w == 0 {
			w = 1
		}
		nd := node{w, n}
		r.nodes = append(r.nodes, nd)
		if i == 0 {
			r.maxWeight = w
			r.gcdWeight = w
			continue
		}
		r.gcdWeight = gcd(w, r.gcdWeight)
		r.maxWeight = max(w, r.maxWeight)

	}
	return r
}

func gcd(a, b int) int {
	c := a % b
	if c == 0 {
		return b
	}
	return gcd(b, c)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
