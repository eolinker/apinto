package discovery

func CheckSkill(skill string) bool {
	return skill == "github.com/eolinker/goku-eosc/discovery.discovery.IDiscovery"
}

//IDiscovery 服务发现接口
type IDiscovery interface {
	GetApp(config string) (IApp, error)
}

//IApp app接口
type IApp interface {
	IAttributes
	ID() string
	Nodes() []INode
	Reset([]INode)
	NodeError(id string) error
	Close() error
}

//IAppContainer app容器接口
type IAppContainer interface {
	Remove(id string) error
}

//INode 节点接口
type INode interface {
	IAttributes
	ID() string
	Ip() string
	Port() int
	Addr() string
	Status() NodeStatus
	Up()
	Down()
	Leave()
}

//Attrs 属性集合
type Attrs map[string]string

//IAttributes 属性接口
type IAttributes interface {
	GetAttrs() Attrs
	GetAttrByName(name string) (string, bool)
}

type NodeStatus int

const (
	Running NodeStatus = 1
	Down    NodeStatus = 2
	Leave   NodeStatus = 3
)
