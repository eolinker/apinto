package store_memory

import (
	"github.com/eolinker/eosc"
)

//Register 注册存储器工厂
func Register() {
	eosc.RegisterStoreDriver("memory", new(Factory))
}

//Factory 存储器工厂结构体
type Factory struct {
}

//Create 创建存储器
func (f *Factory) Create(params map[string]string) (eosc.IStore, error) {

	return NewStore()

}
