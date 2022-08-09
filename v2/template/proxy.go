package template

import (
	"github.com/eolinker/apinto/plugin"
	eoscContext "github.com/eolinker/eosc/eocontext"
	"sync"
)

var _ iProxyDatas = (*ProxyDatas)(nil)

type Proxy struct {
	eoscContext.IChain
	id  string
	org map[string]*plugin.Config

	parent iProxyDatas
}

func (p *Proxy) Destroy() {
	parent := p.parent
	if parent != nil {
		p.parent = nil
		parent.Del(p.id)
	}
	p.IChain.Destroy()
}

type iProxyDatas interface {
	Set(id string, plugins map[string]*plugin.Config) eoscContext.IChain
	Del(id string)
}

type ProxyDatas struct {
	lock    sync.RWMutex
	datas   map[string]*Proxy
	plugins map[string]*plugin.Config
}

func (p *ProxyDatas) Set(id string, conf map[string]*plugin.Config) eoscContext.IChain {
	p.lock.Lock()
	defer p.lock.Unlock()
	cf := plugin.MergeConfig(conf, p.plugins)
	filters := pluginManger.CreateRequest(id, cf)
	v, has := p.datas[id]
	if !has {
		v = &Proxy{
			IChain: filters,
			id:     id,
			org:    conf,
			parent: p,
		}
		p.datas[id] = v
	} else {
		v.IChain = filters
		v.org = conf
	}

	return v

}

func (p *ProxyDatas) Del(id string) {
	p.lock.Lock()
	defer p.lock.Unlock()
	delete(p.datas, id)
}
func (p *ProxyDatas) Reset(conf map[string]*plugin.Config) {
	p.lock.RLock()
	defer p.lock.RUnlock()
	p.plugins = conf
	for _, proxy := range p.datas {
		cf := plugin.MergeConfig(proxy.org, conf)
		proxy.IChain = pluginManger.CreateRequest(proxy.id, cf)
	}
}
func (p *ProxyDatas) Destroy() {
	p.lock.Lock()
	data := p.datas
	p.datas = make(map[string]*Proxy)
	p.lock.Unlock()
	for _, proxy := range data {
		proxy.parent = nil
		proxy.IChain.Destroy()
	}
}
func NewProxyDatas() *ProxyDatas {
	return &ProxyDatas{
		datas: make(map[string]*Proxy),
	}
}
