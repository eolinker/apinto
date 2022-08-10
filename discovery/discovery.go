package discovery

import (
	"errors"
	"github.com/eolinker/eosc/eocontext"
)

var (
	ErrDiscoveryDown = errors.New("discovery down")
)

//CheckSkill 检查目标技能是否符合
func CheckSkill(skill string) bool {
	return skill == "github.com/eolinker/apinto/discovery.discovery.IDiscovery"
}

//IDiscovery 服务发现接口
type IDiscovery interface {
	GetApp(config string) (IApp, error)
}

//IApp app接口
type IApp interface {
	eocontext.EoApp
	IAttributes
	ID() string
	Nodes() []INode
	Reset(nodes Nodes)
	NodeError(id string) error
	Close() error
}

//IAppContainer app容器接口
type IAppContainer interface {
	Remove(id string) error
}

//INode 节点接口
type INode = eocontext.INode

//Attrs 属性集合
type Attrs = eocontext.Attrs

//IAttributes 属性接口
type IAttributes = eocontext.IAttributes

//NodeStatus 节点状态类型
type NodeStatus = eocontext.NodeStatus

const (
	//Running 节点运行中状态
	Running = eocontext.Running
	//Down 节点不可用状态
	Down = eocontext.Down
	//Leave 节点离开状态
	Leave = eocontext.Leave
)
