package router

import (
	"fmt"

	"github.com/eolinker/eosc"
)

//Factory 路由工厂，实现IDriverFactory方法
type Factory struct {
	id         string
	name       string
	group      string
	profession string
	version    string
}

func (f *Factory) Profession() string {
	return f.profession
}

//ID 获取工厂ID
func (f *Factory) ID() string {
	return f.id
}

//Name 获取工厂名称
func (f *Factory) Name() string {
	return f.name
}

//Group 获取工厂分组
func (f *Factory) Group() string {
	return f.group
}

//Version 获取版本号
func (f *Factory) Version() string {
	return f.version
}

func Register() error {
	return nil
}

func NewFactory() *Factory {
	return &Factory{
		id:         fmt.Sprintf("%s:%s_%s:%s", group, profession, name, version),
		name:       name,
		group:      group,
		profession: profession,
		version:    version,
	}
}

func (f *Factory) Create(name string) (eosc.IProfessionDriver, error) {
	r := NewRouter(name)
	return r, nil
}
