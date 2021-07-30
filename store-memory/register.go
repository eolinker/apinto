package store_memory

import (
	"github.com/eolinker/eosc"
)

func Register()  {
	eosc.RegisterStoreDriver("memory",new(Factory))
}

type Factory struct {

}

func (f *Factory) Create(params map[string]string) (eosc.IStore, error) {

	return NewStore()

}
