package ip_hash

import (
	"errors"
	"github.com/eolinker/apinto/discovery"
	"github.com/eolinker/apinto/upstream/balance"
	eoscContext "github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"hash/crc32"
)

const (
	name = "ip-hash"
)

var (
	errNoValidNode = errors.New("no valid node")
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

type ipHash struct {
	app discovery.IApp
}

func (r *ipHash) Select(ctx eoscContext.EoContext) (eoscContext.INode, error) {
	return r.Next(ctx)
}

// Next 由现有节点根据ip_hash决策出一个可用节点
func (r *ipHash) Next(org eoscContext.EoContext) (discovery.INode, error) {
	httpContext, err := http_service.Assert(org)
	if err != nil {
		return nil, err
	}
	readIp := httpContext.Request().ReadIP()
	nodes := r.app.Nodes()
	size := len(nodes)
	if size < 1 {
		return nil, errNoValidNode
	}
	return nodes[HashCode(readIp)%size], nil
}

func newIpHash(app discovery.IApp) *ipHash {
	r := &ipHash{
		app: app,
	}
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
