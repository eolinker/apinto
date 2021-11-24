package discovery

import (
	"sync"

	"github.com/go-basic/uuid"
)

type app struct {
	id            string
	nodes         map[string]INode
	healthChecker IHealthChecker
	attrs         Attrs
	locker        sync.RWMutex
	container     IAppContainer
}

//Reset 重置app的节点列表
func (s *app) Reset(nodes Nodes) {
	tmp := make(map[string]INode)

	for _, node := range nodes {

		if n, has := s.nodes[node.ID()]; has {
			n.Leave()
		}
		tmp[node.ID()] = node

	}
	s.locker.Lock()
	s.nodes = tmp
	s.locker.Unlock()
}

//GetAttrs 获取app的属性集合
func (s *app) GetAttrs() Attrs {
	return s.attrs
}

//GetAttrByName 通过属性名获取app对应属性
func (s *app) GetAttrByName(name string) (string, bool) {
	attr, ok := s.attrs[name]
	return attr, ok
}

//NewApp 创建服务发现app
func NewApp(checker IHealthChecker, container IAppContainer, attrs Attrs, nodes Nodes) IApp {
	if attrs == nil {
		attrs = make(Attrs)
	}
	return &app{
		attrs:         attrs,
		nodes:         nodes,
		locker:        sync.RWMutex{},
		healthChecker: checker,
		id:            uuid.New(),
		container:     container,
	}
}

//ID 返回app的id
func (s *app) ID() string {
	return s.id
}

//Nodes 将运行中的节点列表返回
func (s *app) Nodes() []INode {
	s.locker.RLock()
	defer s.locker.RUnlock()
	nodes := make([]INode, 0, len(s.nodes))
	for _, node := range s.nodes {
		if node.Status() != Running {
			continue
		}
		nodes = append(nodes, node)
	}
	return nodes
}

//NodeError 定时检查节点，当节点失败时，则返回错误
func (s *app) NodeError(id string) error {
	s.locker.Lock()
	defer s.locker.Unlock()
	if n, ok := s.nodes[id]; ok {
		n.Down()
		if s.healthChecker != nil {
			err := s.healthChecker.AddToCheck(n)
			return err
		}
	}
	return nil
}

//Close 关闭服务发现的app
func (s *app) Close() error {
	//
	s.container.Remove(s.id)
	if s.healthChecker != nil {
		return s.healthChecker.Stop()
	}
	return nil
}
