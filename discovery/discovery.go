package discovery

import (
	"errors"
	"github.com/eolinker/eosc/eocontext"
	"sync"
)

var (
	ErrDiscoveryDown = errors.New("discovery down")
)

// IDiscovery 服务发现接口
type IDiscovery interface {
	GetApp(config string) (IAppAgent, error)
}

// CheckSkill 检查目标技能是否符合
func CheckSkill(skill string) bool {
	return skill == "github.com/eolinker/apinto/discovery.discovery.IDiscovery"
}

type NodeInfo struct {
	Ip     string
	Port   int
	Labels map[string]string
}

type IAppContainer interface {
	Set(id string, info []NodeInfo) (app IAppAgent)
	Remove(id string)
	Reset(info map[string][]NodeInfo)
	GetApp(id string) (IAppAgent, bool)
	Keys() []string
}

type appContainer struct {
	nodes nodes
	lock  sync.RWMutex
	apps  map[string]IAppAgent
}

func NewAppContainer() IAppContainer {

	return &appContainer{}
}

func (ac *appContainer) Keys() []string {

	ac.lock.RLock()
	defer ac.lock.RUnlock()
	keys := make([]string, 0, len(ac.apps))
	for k := range ac.apps {
		keys = append(keys, k)
	}
	return keys
}

func (ac *appContainer) create(infos []NodeInfo) []eocontext.INode {
	ns := make([]eocontext.INode, 0, len(infos))
	for _, i := range infos {

		n := ac.nodes.Get(i.Ip, i.Port)
		ns = append(ns, NewNode(n, i.Labels))
	}
	return ns
}
func (ac *appContainer) Set(name string, infos []NodeInfo) IAppAgent {

	ns := ac.create(infos)
	ac.lock.RLock()
	app, has := ac.apps[name]
	ac.lock.Unlock()
	if has {
		app.reset(ns)
		return app
	}

	ac.lock.Lock()
	app, has = ac.apps[name]

	if !has {
		app = newApp(ns)
		ac.apps[name] = app
	}
	ac.lock.Unlock()

	return app
}

func (ac *appContainer) Remove(name string) {
	ac.lock.Lock()
	defer ac.lock.RUnlock()
	delete(ac.apps, name)
}

func (ac *appContainer) Reset(infos map[string][]NodeInfo) {
	nm := make(map[string]IAppAgent)
	for name, info := range infos {
		nm[name] = newApp(ac.create(info))
	}
	ac.lock.Lock()
	ac.apps = nm
	ac.lock.Unlock()

}

func (ac *appContainer) GetApp(name string) (IAppAgent, bool) {
	ac.lock.RLock()
	defer ac.lock.RUnlock()
	app, has := ac.apps[name]

	return app, has

}
