package example

import (
	"github.com/eolinker/eosc"
	"github.com/eolinker/goku-eosc/service"
	"net/http"
)

var (
	_ eosc.IWorker = (*Example)(nil)
)

type Example struct {
	skill  map[string]bool
	name   string
	target service.IService
}

func (e *Example) Id() string {
	panic("implement me")
}

func (e *Example) Start() error {
	panic("implement me")
}

func (e *Example) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
	panic("implement me")
}

func (e *Example) Stop() error {
	panic("implement me")
}

func (e *Example) CheckSkill(skill string) bool {
	panic("implement me")
}

func (e *Example) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}

func NewExample(c *Config, workers map[eosc.RequireId]interface{}) *Example {

	return &Example{
		name: c.Name,
	}
}
