package round_robin

import (
	"errors"
	"github.com/eolinker/apinto/discovery"
	"github.com/eolinker/apinto/upstream/balance"
	eoscContext "github.com/eolinker/eosc/eocontext"
	"strconv"
)

const (
	name = "round-robin"
)

var (
	errNoValidNode = errors.New("no valid node")
)

type roundRobinKeyType struct {
}

var (
	roundRobinKey = roundRobinKeyType{}
)

// Register 注册round-robin算法
func Register() {
	balance.Register(name, newRoundRobinFactory())
}

func newRoundRobinFactory() *roundRobinFactory {
	return &roundRobinFactory{}
}

type roundRobinFactory struct {
}

// Create 创建一个round-Robin算法处理器
func (r *roundRobinFactory) Create() (eoscContext.BalanceHandler, error) {
	rr := newRoundRobin()
	return rr, nil
}

type node struct {
	weight int
	eoscContext.INode
}

type roundRobin struct {
	index int
}
type roundRobinContext struct {
	nodes []node
	// gcdWeight 权重最大公约数
	gcdWeight int
	// maxWeight 权重最大值
	maxWeight int

	cw        int
	lastIndex int
}

func (r *roundRobin) Select(ctx eoscContext.EoContext) (eoscContext.INode, int, error) {
	return r.Next(ctx)
}

// Next 由现有节点根据round_Robin决策出一个可用节点
func (r *roundRobin) Next(ctx eoscContext.EoContext) (eoscContext.INode, int, error) {

	rc := r.init(ctx)
	size := len(rc.nodes)
	if size < 1 {
		return nil, 0, errNoValidNode
	}
	if rc.lastIndex < 0 {
		rc.lastIndex = r.index
	}
	for i := 0; i < size; i++ {

		index := rc.lastIndex
		rc.lastIndex++

		index %= size
		if index == 0 {
			rc.cw = rc.cw - rc.gcdWeight
			if rc.cw <= 0 {
				rc.cw = rc.maxWeight
				if rc.cw == 0 {
					return nil, 0, errNoValidNode
				}
			}
		}

		if rc.nodes[index].weight >= rc.cw {
			if rc.nodes[index].Status() == discovery.Down {

				continue
			}
			return rc.nodes[index], index, nil
		}

	}
	return nil, 0, errNoValidNode
}

func (r *roundRobin) init(ctx eoscContext.EoContext) *roundRobinContext {

	nodesValue := ctx.Value(roundRobinKey)
	if nodesValue != nil {
		return nodesValue.(*roundRobinContext)
	}

	nodes := ctx.GetApp().Nodes()
	rc := create(nodes)
	ctx.WithValue(roundRobinKey, rc)

	return rc

}
func create(nodes []eoscContext.INode) *roundRobinContext {
	rc := new(roundRobinContext)
	rc.lastIndex = -1
	rc.nodes = make([]node, 0, len(nodes))
	for i, n := range nodes {

		weight, _ := n.GetAttrByName("weight")
		w, _ := strconv.Atoi(weight)
		if w == 0 {
			w = 1
		}
		nd := node{w, n}
		rc.nodes = append(rc.nodes, nd)
		if i == 0 {
			rc.maxWeight = w
			rc.gcdWeight = w
			continue
		}
		rc.gcdWeight = gcd(w, rc.gcdWeight)
		rc.maxWeight = max(w, rc.maxWeight)
	}
	return rc
}
func newRoundRobin() *roundRobin {
	r := &roundRobin{}

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
