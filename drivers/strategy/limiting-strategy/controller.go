package limiting_strategy

import (
	"github.com/eolinker/eosc"
	"reflect"
)

var (
	controller               = NewController()
	_          eosc.ISetting = controller
	_          IController   = controller
)

type IController interface {
	Store(id string)
	Del(id string)
}
type Controller struct {
	profession string
	driver     string
	all        map[string]struct{}
}

func (c *Controller) Store(id string) {
	c.all[id] = struct{}{}
}

func (c *Controller) Del(id string) {
	delete(c.all, id)
}

func (c *Controller) ConfigType() reflect.Type {
	return configType
}

func (c *Controller) Set(conf interface{}) (err error) {
	return eosc.ErrorUnsupportedKind
}

func (c *Controller) Get() interface{} {
	return nil
}

func (c *Controller) Mode() eosc.SettingMode {
	return eosc.SettingModeBatch
}

func (c *Controller) Check(cfg interface{}) (profession, name, driver, desc string, err error) {
	conf, ok := cfg.(*Config)
	if !ok {
		err = eosc.ErrorConfigIsNil
		return
	}
	if empty(conf.Name) {
		err = eosc.ErrorConfigFieldUnknown
		return
	}
	err = checkConfig(conf)
	if err != nil {
		return
	}
	return c.profession, conf.Name, c.driver, conf.Description, nil

}
func empty(vs ...string) bool {
	for _, v := range vs {
		if len(v) == 0 {
			return true
		}
	}
	return false
}
func (c *Controller) AllWorkers() []string {
	ws := make([]string, 0, len(c.all))
	for id := range c.all {
		ws = append(ws, id)
	}
	return ws
}

func NewController() *Controller {
	return &Controller{
		all: map[string]struct{}{},
	}
}
