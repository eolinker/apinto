package router_http

import (
	"strconv"

	"github.com/eolinker/eosc"
)

var _ IRouters = (*Routers)(nil)

//IRouters 路由树管理器实现的接口
type IRouters interface {
	Set(port int, id string, conf *Config) (IRouter, bool, error)
	SetAll(id string, conf *Config) (map[int]IRouter, error)
	Del(port int, id string) (IRouter, bool)
}

//Routers 路由树管理器的结构体
type Routers struct {
	data eosc.IUntyped
}

func (rs *Routers) SetAll(id string, conf *Config) (map[int]IRouter, error) {
	routers := make(map[int]IRouter)
	for key, r := range rs.data.All() {
		router := r.(IRouter)
		err := router.SetRouter(id, conf)
		if err != nil {
			return nil, err
		}
		port, err := strconv.Atoi(key)
		if err != nil {
			return nil, err
		}
		routers[port] = router
	}
	return routers, nil
}

//Set 将路由配置加入到对应端口的路由树中
func (rs *Routers) Set(port int, id string, conf *Config) (IRouter, bool, error) {
	name := strconv.Itoa(port)
	r, has := rs.data.Get(name)

	//若对应端口不存在路由树，则新建
	if !has {
		return nil, false, nil
		//router := NewRouter()
		//err := router.SetRouter(id, conf)
		//if err != nil {
		//	return nil, false, err
		//}
		//rs.data.Set(name, router)
		//return router, true, nil
	}
	// todo 这里需要校验端口已使用的的http协议是否与之前配置冲突，并返回新的合并后的证书列表

	router := r.(IRouter)
	err := router.SetRouter(id, conf)
	if err != nil {
		return nil, false, err
	}
	return router, false, nil
}

//NewRouters 新建路由树管理器
func NewRouters() *Routers {
	return &Routers{
		data: eosc.NewUntyped(),
	}
}

//func (rs *Routers) GetEmployee(port int) (IRouter, bool) {
//	name := strconv.Itoa(port)
//	r, has := rs.data.GetEmployee(name)
//	if !has {
//		var router IRouter = NewRouter()
//		rs.data.SetStatus(name, router)
//		return router, true
//	}
//	return r.(IRouter), false
//}

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
