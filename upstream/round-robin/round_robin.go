package round_robin

import (
	"errors"
	"github.com/eolinker/apinto/utils/queue"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	eoscContext "github.com/eolinker/eosc/eocontext"

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
func noValidNodeHandler() (eoscContext.INode, int, error) {
	return nil, 0, errNoValidNode
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
	index           int
	weight          int64
	effectiveWeight int64
	node            eoscContext.INode
}

type roundRobin struct {
	eoscContext.EoApp
	scheme  string
	timeout time.Duration

	// 节点数量,也是每次最大遍历数
	size int

	// gcdWeight 权重最大公约数
	gcdWeight int64

	nextHandler func() (eoscContext.INode, int, error)

	locker        sync.Mutex
	nodeQueueNext queue.Queue[node]
	nodeQueue     queue.Queue[node]
	updateTime    int64
}

func (r *roundRobin) Scheme() string {
	return r.scheme
}

func (r *roundRobin) TimeOut() time.Duration {
	return r.timeout
}

func (r *roundRobin) Select(ctx eoscContext.EoContext) (eoscContext.INode, int, error) {
	r.tryReset()
	if r.nextHandler != nil {
		return r.nextHandler()
	}

	return r.Next()
}

// Next 由现有节点根据round_Robin决策出一个可用节点
func (r *roundRobin) Next() (eoscContext.INode, int, error) {

	r.locker.Lock()
	defer r.locker.Unlock()
	for i := 0; i < r.size; i++ {

		if r.nodeQueue.Empty() {
			r.nodeQueue, r.nodeQueueNext = r.nodeQueueNext, r.nodeQueue
		}
		entry := r.nodeQueue.Pop()
		nodeValue := entry.Value()

		nodeValue.effectiveWeight -= r.gcdWeight

		if nodeValue.weight > 0 {
			r.nodeQueue.Push(entry)
		} else {
			nodeValue.effectiveWeight = nodeValue.weight
			r.nodeQueueNext.Push(entry)
		}
		if nodeValue.node.Status() == eoscContext.Down {
			// 如果节点down( 开启健康检查才会出现down 状态) 则去拿下一个节点
			continue
		}
		return nodeValue.node, nodeValue.index, nil
	}
	return nil, 0, errNoValidNode

}

func (r *roundRobin) tryReset() {
	now := time.Now().Unix()
	if now-atomic.LoadInt64(&r.updateTime) < 30 {
		return
	}
	r.locker.Lock()
	defer r.locker.Unlock()
	if now-atomic.LoadInt64(&r.updateTime) < 30 {
		return
	}
	atomic.StoreInt64(&r.updateTime, now)

	nodes := r.Nodes()
	size := len(nodes)
	if size == 0 {
		r.nextHandler = noValidNodeHandler
		return
	}
	if size == 1 {
		node := nodes[0]
		r.nextHandler = func() (eoscContext.INode, int, error) {
			return node, 0, nil
		}
		return
	}

	ns := make([]*node, 0, size)
	gcdWeight := int64(0)
	for _, n := range nodes {

		weight, _ := n.GetAttrByName("weight")
		w, _ := strconv.ParseInt(weight, 10, 64)
		if w == 0 {
			w = 1
		}
		nd := &node{
			weight: w, effectiveWeight: w,
			node: n,
		}
		ns = append(ns, nd)

		gcdWeight = gcd(w, gcdWeight) // 计算权重的最大公约数
	}
	r.size = size

	r.gcdWeight = gcdWeight
	r.nodeQueue = queue.NewQueue(ns...)
	r.nodeQueueNext = queue.NewQueue[node]()
	r.nextHandler = nil
}

func newRoundRobin(app eoscContext.EoApp, scheme string, timeout time.Duration) *roundRobin {
	r := &roundRobin{
		EoApp:   app,
		scheme:  scheme,
		timeout: timeout,
	}

	return r
}

type intType interface {
	int | int64 | int32 | int16 | int8 | uint64 | uint32 | uint16 | uint8 | uint
}

func gcd[T intType](a, b T) T {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

func max[T intType](a, b T) T {
	if a > b {
		return a
	}
	return b
}
