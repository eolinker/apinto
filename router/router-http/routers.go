package router_http

import (
	"strconv"

	"github.com/eolinker/goku/plugin"

	"github.com/eolinker/eosc"
)

var _ IRouters = (*Routers)(nil)

//IRouters 路由树管理器实现的接口
type IRouters interface {
	Set(port int, id string, conf *Config) (IRouter, bool, error)
	Del(port int, id string) (IRouter, bool)
}

//Routers 路由树管理器的结构体
type Routers struct {
	data          eosc.IUntyped
	pluginManager plugin.IPluginManager
}

//Set 将路由配置加入到对应端口的路由树中
func (rs *Routers) Set(port int, id string, conf *Config) (IRouter, bool, error) {
	name := strconv.Itoa(port)
	r, has := rs.data.Get(name)

	//若对应端口不存在路由树，则新建
	if !has {
		globalRouterFilter := rs.pluginManager.CreateRequest(name, map[string]*plugin.Config{})
		router := NewRouter(globalRouterFilter)
		err := router.SetRouter(id, conf)
		if err != nil {
			return nil, false, err
		}
		rs.data.Set(name, router)
		return router, true, nil
	}

	router := r.(IRouter)
	err := router.SetRouter(id, conf)
	if err != nil {
		return nil, false, err
	}
	return router, false, nil
}

//NewRouters 新建路由树管理器
func NewRouters(pluginManager plugin.IPluginManager) *Routers {
	rs := &Routers{
		data:          eosc.NewUntyped(),
		pluginManager: pluginManager,
	}

	return rs
}

//Del 将路由配置从对应端口的路由树中删去
func (rs *Routers) Del(port int, id string) (IRouter, bool) {
	name := strconv.Itoa(port)
	if i, has := rs.data.Get(name); has {
		r := i.(IRouter)
		count := r.Del(id)
		if count == 0 {
			rs.data.Del(name)
		}
		return r, true
	}
	return nil, false

}
