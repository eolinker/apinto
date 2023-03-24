package discovery_old

import (
	"errors"
	"github.com/eolinker/eosc/eocontext"
)

var (
	ErrDiscoveryDown = errors.New("discovery down")
)

// CheckSkill 检查目标技能是否符合
func CheckSkill(skill string) bool {
	return skill == "github.com/eolinker/apinto/discovery.discovery.IDiscovery"
}

// IDiscovery 服务发现接口
type IDiscovery interface {
	GetApp(config string) (IApp, error)
}

// IApp app接口
type IApp interface {
	IAttributes
	ID() string
	Nodes() []eocontext.INode
	Reset(nodes Nodes)
	NodeError(id string) error
	Close() error
}

type INode = eocontext.INode

// BaseNode 节点接口
type BaseNode interface {
	ID() string
	IP() string
	Port() int
	Addr() string
	Status() NodeStatus
	Up()
	Down()
	Leave()
}

// Attrs 属性集合
type Attrs = eocontext.Attrs

// IAttributes 属性接口
type IAttributes = eocontext.IAttributes

// NodeStatus 节点状态类型
type NodeStatus = eocontext.NodeStatus

const (
	//Running 节点运行中状态
	Running = eocontext.Running
	//Down 节点不可用状态
	Down = eocontext.Down
	//Leave 节点离开状态
	Leave = eocontext.Leave
)
