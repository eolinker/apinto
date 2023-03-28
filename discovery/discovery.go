package discovery

import (
	"errors"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrDiscoveryDown = errors.New("discovery down")
)

// IDiscovery 服务发现接口
type IDiscovery interface {
	GetApp(config string) (IApp, error)
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
	INodes
	Set(id string, info []NodeInfo) (app IAppAgent)
	Reset(info map[string][]NodeInfo)
	GetApp(id string) (IAppAgent, bool)
	Keys() []string
}

type appContainer struct {
	lock          sync.RWMutex
	nodes         eosc.Untyped[string, INode]
	apps          map[string]*_AppAgent
	isHealthCheck int32
}

func (ac *appContainer) status(status NodeStatus) NodeStatus {

	if atomic.LoadInt32(&ac.isHealthCheck) > 0 {
		return status
	}
	return Running
}

func (ac *appContainer) SetHealthCheck(isHealthCheck bool) {
	if isHealthCheck {
		atomic.StoreInt32(&ac.isHealthCheck, 1)
	} else {
		atomic.StoreInt32(&ac.isHealthCheck, 0)

	}
}

func NewAppContainer() IAppContainer {

	return &appContainer{
		apps:  make(map[string]*_AppAgent),
		nodes: eosc.BuildUntyped[string, INode](),
	}
}

func (ac *appContainer) Keys() []string {

	ac.lock.RLock()
	defer ac.lock.RUnlock()
	if ac.apps == nil {
		return nil
	}
	keys := make([]string, 0, len(ac.apps))
	for k := range ac.apps {
		keys = append(keys, k)
	}
	return keys
}

func (ac *appContainer) create(infos []NodeInfo) []eocontext.INode {
	ns := make([]eocontext.INode, 0, len(infos))
	for _, i := range infos {

		n := ac.Get(i.Ip, i.Port)
		ns = append(ns, NewNode(n, i.Labels))
	}
	return ns
}
func (ac *appContainer) Set(name string, infos []NodeInfo) IAppAgent {

	ns := ac.create(infos)
	ac.lock.RLock()
	app, has := ac.apps[name]
	ac.lock.RUnlock()
	if has {
		app.reset(ns)
		return app
	}

	ac.lock.Lock()

	app, has = ac.apps[name]
	needCheck := false
	if !has {
		if len(ac.apps) == 0 {
			needCheck = true
		}
		app = newApp(ns)
		ac.apps[name] = app
	}
	ac.lock.Unlock()
	if needCheck {
		go ac.doCheck()
	}
	return app
}
func (ac *appContainer) doCheck() {
	t := time.NewTicker(time.Second * 10)
	defer t.Stop()
	for range t.C {

		ac.lock.Lock()
		if len(ac.apps) == 0 {
			ac.lock.Unlock()
			return
		}
		for key, app := range ac.apps {
			if atomic.LoadInt64(&app.use) <= 0 {
				delete(ac.apps, key)
			}
		}

		nodeUse := make(map[string]int)

		for _, app := range ac.apps {
			for _, n := range app.nodes {
				nodeUse[n.ID()] += 1
			}
		}
		nodeList := ac.nodes.List()
		for _, n := range nodeList {
			if nodeUse[n.ID()] == 0 {
				ac.nodes.Del(n.ID())
			}
		}
		ac.lock.Unlock()
	}

}

func (ac *appContainer) Reset(infos map[string][]NodeInfo) {
	nm := make(map[string]*_AppAgent)
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
