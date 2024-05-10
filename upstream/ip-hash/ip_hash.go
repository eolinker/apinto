package ip_hash

import (
	"errors"
	"hash/crc32"
	"time"

	"github.com/eolinker/apinto/upstream/balance"
	eoscContext "github.com/eolinker/eosc/eocontext"
)

const (
	name = "ip-hash"
)

var (
	errNoValidNode                            = errors.New("no valid node")
	_              eoscContext.BalanceHandler = (*ipHash)(nil)
	_              balance.IBalanceFactory    = (*ipHashFactory)(nil)
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
func (r *ipHashFactory) Create(app eoscContext.EoApp, scheme string, timeout time.Duration) (eoscContext.BalanceHandler, error) {
	rr := newIpHash(app, scheme, timeout)
	return rr, nil
}

type ipHash struct {
	eoscContext.EoApp
	scheme  string
	timeout time.Duration
}

func (r *ipHash) Scheme() string {
	return r.scheme
}

func (r *ipHash) TimeOut() time.Duration {
	return r.timeout
}

func (r *ipHash) Select(ctx eoscContext.EoContext) (eoscContext.INode, int, error) {
	return r.Next(ctx)
}

// Next 由现有节点根据ip_hash决策出一个可用节点
func (r *ipHash) Next(org eoscContext.EoContext) (eoscContext.INode, int, error) {

	readIp := org.RealIP()
	nodes := r.Nodes()
	size := len(nodes)
	if size == 1 {
		return nodes[0], 0, nil
	}
	if size < 1 {
		return nil, 0, errNoValidNode
	}
	index := HashCode(readIp) % size
	return nodes[index], index, nil
}

func newIpHash(app eoscContext.EoApp, scheme string, timeout time.Duration) *ipHash {
	r := &ipHash{EoApp: app, scheme: scheme, timeout: timeout}
	return r
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
