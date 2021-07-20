package http_router

import (
	"github.com/eolinker/eosc"
)

type Router struct {
}

func (r *Router) Marshal() ([]byte, error) {
	panic("implement me")
}

func (r *Router) Worker() (eosc.IWorker, error) {
	panic("implement me")
}

func (r *Router) CheckSkill(skill string) bool {
	panic("implement me")
}

func (r *Router) Info() eosc.WorkerInfo {
	return eosc.WorkerInfo{
		Id:     "",
		Name:   "",
		Driver: "",
		Create: "",
		Update: "",
	}
}

func NewRouter(c *Config) *Router {
	return &Router{}
}
