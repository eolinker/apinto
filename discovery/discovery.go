package discovery

//CheckSkill 检查目标技能是否符合
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
	IP() string
	Port() int
	Scheme() string
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

//NodeStatus 节点状态类型
type NodeStatus int

const (
	//Running 节点运行中状态
	Running NodeStatus = 1
	//Down 节点不可用状态
	Down NodeStatus = 2
	//Leave 节点离开状态
	Leave NodeStatus = 3
)
