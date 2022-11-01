package ip_hash

import (
	"errors"
	"github.com/eolinker/apinto/discovery"
	"github.com/eolinker/apinto/upstream/balance"
	eoscContext "github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"hash/crc32"
	"time"
)

const (
	name = "ip-hash"
)

var (
	errNoReadIp = errors.New("read ip is null")
)

// Register 注册ip-hash算法
func Register() {
	balance.Register(name, newIpHashFactory())
}

func newIpHashFactory() *ipHashFactory {
	return &ipHashFactory{}
}

type ipHashFactory struct {
}

// Create 创建一个ip-hash算法处理器
func (r *ipHashFactory) Create(app discovery.IApp) (eoscContext.BalanceHandler, error) {
	rr := newIpHash(app)
	return rr, nil
}

type node struct {
	weight int
	discovery.INode
}

type ipHash struct {
	app discovery.IApp
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

	downNodes map[int]discovery.INode
}

func (r *ipHash) Select(ctx eoscContext.EoContext) (eoscContext.INode, error) {
	return r.Next(ctx)
}

// Next 由现有节点根据ip_hash决策出一个可用节点
func (r *ipHash) Next(org eoscContext.EoContext) (discovery.INode, error) {
	if time.Now().Sub(r.updateTime) > time.Second*30 {
		// 当上次节点更新时间与当前时间间隔超过30s，则重新设置节点
		r.set()
	}
	ctx, err := http_service.Assert(org)
	if err != nil {
		return nil, err
	}
	readIp := ctx.Request().ReadIP()
	if len(readIp) == 0 {
		return nil, errNoReadIp
	}
	ipHash := HashCode(readIp)
	index := ipHash % r.size
	return r.nodes[index], nil
}

func newIpHash(app discovery.IApp) *ipHash {
	r := &ipHash{
		app: app,
	}
	r.set()
	return r
}

func (r *ipHash) set() {
	r.downNodes = make(map[int]discovery.INode)
	nodes := r.app.Nodes()
	r.size = len(nodes)
	ns := make([]node, 0, r.size)
	for _, n := range nodes {
		nd := node{1, n}
		ns = append(ns, nd)
	}
	r.nodes = ns
	r.updateTime = time.Now()
}

func HashCode(s string) int {
	v := int(crc32.ChecksumIEEE([]byte(s)))
	if v >= 0 {
		return v
	}
	if -v >= 0 {
		return -v
	}
	// v == MinInt
	return 0
}
