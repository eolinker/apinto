package router

import (
	"github.com/eolinker/eosc"
	"net/http"
	"sync"
)

var _ IRouter = (*Router)(nil)

type IRouter interface {
	SetRouter(id string, config *Config) error
	Count() int
	Del(id string) int
	http.Handler
}

type Router struct {
	locker sync.Locker
	data   eosc.IUntyped
	match  IMatcher
}

func NewRouter() *Router {
	return &Router{
		locker: &sync.Mutex{},
	}
}

func (r *Router) Count() int {
	return r.data.Count()
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	h, has := r.match.Match(req)
	if !has {
		http.NotFound(w, req)
		return
	}

}

func (r *Router) SetRouter(id string, config *Config) error {
	r.locker.Lock()
	defer r.locker.Unlock()
	data := r.data.Clone()
	data.Set(id, config)
	list := data.List()
	cs := make([]*Config, 0, len(list))
	for _, i := range list {
		cs = append(cs, i.(*Config))
	}
	matcher, err := parse(cs)
	if err != nil {
		return err
	}
	r.match = matcher
	r.data = data
	return nil
}

func (r *Router) Del(id string) int {
	r.locker.Lock()
	defer r.locker.Unlock()

	data := r.data.Clone()
	data.Del(id)
	if data.Count() == 0 {
		r.match = nil
	} else {
		list := data.List()
		cs := make([]*Config, 0, len(list))
		for _, i := range list {
			cs = append(cs, i.(*Config))
		}
		m, err := parse(cs)
		if err != nil {
			return r.data.Count()
		}
		r.match = m
	}

	return r.data.Count()
}
