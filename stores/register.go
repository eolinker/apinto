package stores

import "github.com/eolinker/eosc"

var (
	defaultDriverRegister iDriverRegister = newDriverManager()
)

type iDriverRegister interface {
	RegisterDriverByKey(key string, factory IStoresFactory)
	GetDriverByKey(key string) (IStoresFactory, bool)
	Keys() []string
}

type DriverRegister struct {
	register eosc.IRegister
	keys     []string
}

func newDriverManager() *DriverRegister {
	return &DriverRegister{
		register: eosc.NewRegister(),
		keys:     make([]string, 0, 10),
	}
}

func (dm *DriverRegister) GetDriverByKey(key string) (IStoresFactory, bool) {
	o, has := dm.register.Get(key)
	if has {
		f, ok := o.(IStoresFactory)
		return f, ok
	}
	return nil, false
}

func (dm *DriverRegister) RegisterDriverByKey(key string, factory IStoresFactory) {
	dm.register.Register(key, factory, true)
	dm.keys = append(dm.keys, key)
}

func (dm *DriverRegister) Keys() []string {
	return dm.keys
}

func RegisterFactory(key string, factory IStoresFactory) {
	defaultDriverRegister.RegisterDriverByKey(key, factory)
}

func Get(key string) (IStoresFactory, bool) {
	return defaultDriverRegister.GetDriverByKey(key)
}

func Keys() []string {
	return defaultDriverRegister.Keys()
}
