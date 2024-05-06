package round_robin

import (
	"errors"
	"strconv"
	"time"

	eoscContext "github.com/eolinker/eosc/eocontext"

	"github.com/eolinker/apinto/discovery"
	"github.com/eolinker/apinto/upstream/balance"
)

const (
	name = "round-robin"
)

var (
	errNoValidNode                            = errors.New("no valid node")
	_              eoscContext.BalanceHandler = (*roundRobin)(nil)
	_              balance.IBalanceFactory    = (*roundRobinFactory)(nil)
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

func (r roundRobinFactory) Create(app eoscContext.EoApp, scheme string, timeout time.Duration) (eoscContext.BalanceHandler, error) {
	rr := newRoundRobin(app, scheme, timeout)
	return rr, nil
}

type node struct {
	weight int
	eoscContext.INode
}

type roundRobin struct {
	eoscContext.EoApp
	scheme  string
	timeout time.Duration
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

	cw int

	updateTime time.Time

	downNodes map[int]eoscContext.INode
}

func (r *roundRobin) Scheme() string {
	return r.scheme
}

func (r *roundRobin) TimeOut() time.Duration {
	return r.timeout
}

func (r *roundRobin) Select(ctx eoscContext.EoContext) (eoscContext.INode, int, error) {
	return r.Next()
}

// Next 由现有节点根据round_Robin决策出一个可用节点
func (r *roundRobin) Next() (eoscContext.INode, int, error) {
	if time.Now().Sub(r.updateTime) > time.Second*30 {
		// 当上次节点更新时间与当前时间间隔超过30s，则重新设置节点
		r.set()
	}
	if r.size < 1 {
		return nil, 0, errNoValidNode
	}
	for {
		index := r.index % r.size
		r.index = (r.index + 1) % r.size
		if len(r.downNodes) >= r.size {
			return nil, 0, errNoValidNode
		}

		if index == 0 {
			r.cw = r.cw - r.gcdWeight
			if r.cw <= 0 {
				r.cw = r.maxWeight
				if r.cw == 0 {
					return nil, 0, errNoValidNode
				}
			}
		}

		if r.nodes[index].weight >= r.cw {
			if r.nodes[index].Status() == discovery.Down {
				r.downNodes[index] = r.nodes[index]
				continue
			}
			return r.nodes[index], index, nil
		}
	}
}

func (r *roundRobin) set() {
	r.downNodes = make(map[int]eoscContext.INode)
	nodes := r.Nodes()
	r.size = len(nodes)
	ns := make([]node, 0, r.size)
	for i, n := range nodes {

		weight, _ := n.GetAttrByName("weight")
		w, _ := strconv.Atoi(weight)
		if w == 0 {
			w = 1
		}
		nd := node{w, n}
		ns = append(ns, nd)
		if i == 0 {
			r.maxWeight = w
			r.gcdWeight = w
			continue
		}
		r.gcdWeight = gcd(w, r.gcdWeight)
		r.maxWeight = max(w, r.maxWeight)
	}
	r.nodes = ns
	r.updateTime = time.Now()
}

func newRoundRobin(app eoscContext.EoApp, scheme string, timeout time.Duration) *roundRobin {
	r := &roundRobin{
		EoApp:   app,
		scheme:  scheme,
		timeout: timeout,
	}
	r.set()
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
