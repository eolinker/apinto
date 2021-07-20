package balance

import (
	"errors"

	"github.com/eolinker/eosc"
	"github.com/eolinker/goku-eosc/discovery"
)

var (
	defaultDriverRegister iDriverRegister = newDriverManager()
)

type IBalanceFactory interface {
	Create(app discovery.IApp) (IBalanceHandler, error)
}

type IBalanceHandler interface {
	Next() (discovery.INode, error)
}

type iDriverRegister interface {
	RegisterDriverByKey(key string, factory IBalanceFactory)
	GetDriverByKey(key string) (IBalanceFactory, bool)
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

func (dm *DriverRegister) GetDriverByKey(key string) (IBalanceFactory, bool) {
	o, has := dm.register.Get(key)
	if has {
		f, ok := o.(IBalanceFactory)
		return f, ok
	}
	return nil, false
}

func (dm *DriverRegister) RegisterDriverByKey(key string, factory IBalanceFactory) {
	dm.register.Register(key, factory, true)
	dm.keys = append(dm.keys, key)
}

func (dm *DriverRegister) Keys() []string {
	return dm.keys
}

func Register(key string, factory IBalanceFactory) {
	defaultDriverRegister.RegisterDriverByKey(key, factory)
}

func Get(key string) (IBalanceFactory, bool) {
	return defaultDriverRegister.GetDriverByKey(key)
}

func Keys() []string {
	return defaultDriverRegister.Keys()
}

func GetDriver(name string, app discovery.IApp) (IBalanceHandler, error) {
	factory, ok := Get(name)
	if !ok {
		for _, key := range Keys() {
			factory, ok = Get(key)
			if ok {
				break
			}
		}
		if factory == nil {
			return nil, errors.New("no valid balance handler")
		}
	}
	return factory.Create(app)
}
