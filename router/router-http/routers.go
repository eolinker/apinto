package router_http

import (
	"github.com/eolinker/eosc"
	"strconv"
)

var _ IRouters = (*Routers)(nil)

type IRouters interface {
	Set(port int, id string, conf *Config) (IRouter, bool, error)
 	Del(port int, id string) (IRouter, bool)
}
type Routers struct {
	data eosc.IUntyped
}

func (rs *Routers) Set(port int, id string, conf *Config) (IRouter, bool, error) {
	name := strconv.Itoa(port)
	r, has := rs.data.Get(name)

	if !has {
		router := NewRouter()
		err := router.SetRouter(id, conf)
		if err != nil {
			return nil, false, err
		}
		rs.data.Set(id, router)
		return router, true, nil
	} else {
		router := r.(IRouter)
		err := router.SetRouter(id, conf)
		if err != nil {
			return nil, false, err
		}
		return router, false, nil
	}

}

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
//		rs.data.Set(name, router)
//		return router, true
//	}
//	return r.(IRouter), false
//}

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
