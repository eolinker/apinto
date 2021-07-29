package discovery

import (
	"github.com/eolinker/eosc"
)

type services struct {
	apps        eosc.IUntyped
	appNameOfID eosc.IUntyped
}

//NewServices 创建服务发现的服务app集合
func NewServices() IServices {
	return &services{apps: eosc.NewUntyped(), appNameOfID: eosc.NewUntyped()}
}

//get 获取对应服务名的节点列表
func (s *services) get(serviceName string) (eosc.IUntyped, bool) {
	v, ok := s.apps.Get(serviceName)
	if !ok {
		return nil, ok
	}
	apps, ok := v.(eosc.IUntyped)
	return apps, ok
}

//func (s *services) Get(serviceName string) (IApp, bool) {
//	if apps, ok := s.get(serviceName); ok {
//		for _, r := range apps.List() {
//			v, ok := r.(IApp)
//			if ok {
//				return v, true
//			}
//		}
//	}
//	return nil, false
//}

//Set 将app存入其对应服务的节点列表
func (s *services) Set(serviceName string, id string, app IApp) error {
	s.appNameOfID.Set(id, serviceName)
	if apps, ok := s.get(serviceName); ok {
		apps.Set(id, app)
		return nil
	}
	apps := eosc.NewUntyped()
	apps.Set(id, app)
	s.apps.Set(serviceName, apps)
	return nil
}

//Remove 将目标app从其对应服务的app列表中删除，传入值为目标app的id
func (s *services) Remove(appID string) (string, int) {
	v, has := s.appNameOfID.Del(appID)
	if has {
		name := v.(string)
		apps, ok := s.get(name)
		if ok {
			apps.Del(appID)
			return name, apps.Count()
		}
		return name, 0
	}
	return "", 0
}

//Update 更新目标服务所有app的节点列表
func (s *services) Update(serviceName string, nodes []INode) error {
	if apps, ok := s.get(serviceName); ok {
		for _, r := range apps.List() {
			v, ok := r.(IApp)
			if ok {
				v.Reset(nodes)
			}
		}
	}
	return nil
}

//AppKeys 获取现有服务app的服务名列表
func (s *services) AppKeys() []string {
	return s.apps.Keys()
}

//IServices 服务app集合接口
type IServices interface {
	Set(serviceName string, id string, app IApp) error
	Remove(id string) (string, int)
	Update(serviceName string, nodes Nodes) error
	AppKeys() []string
	//Get(serviceName string) (IApp, bool)
}
