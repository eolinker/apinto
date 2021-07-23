package balance

import (
	"errors"

	"github.com/eolinker/eosc"
	"github.com/eolinker/goku-eosc/discovery"
)

var (
	defaultBalanceFactoryRegister IBalanceFactoryRegister = newBalanceFactoryManager()
)

//IBalanceFactory 实现了负载均衡算法工厂
type IBalanceFactory interface {
	Create(app discovery.IApp) (IBalanceHandler, error)
}

//IBalanceHandler 实现了负载均衡算法
type IBalanceHandler interface {
	Next() (discovery.INode, error)
}

//IBalanceFactoryRegister 实现了负载均衡算法工厂管理器
type IBalanceFactoryRegister interface {
	RegisterFactoryByKey(key string, factory IBalanceFactory)
	GetFactoryByKey(key string) (IBalanceFactory, bool)
	Keys() []string
}

//driverRegister 实现了IBalanceFactoryRegister接口
type driverRegister struct {
	register eosc.IRegister
	keys     []string
}

//newBalanceFactoryManager 创建负载均衡算法工厂管理器
func newBalanceFactoryManager() IBalanceFactoryRegister {
	return &driverRegister{
		register: eosc.NewRegister(),
		keys:     make([]string, 0, 10),
	}
}

//GetFactoryByKey 获取指定balance工厂
func (dm *driverRegister) GetFactoryByKey(key string) (IBalanceFactory, bool) {
	o, has := dm.register.Get(key)
	if has {
		f, ok := o.(IBalanceFactory)
		return f, ok
	}
	return nil, false
}

//RegisterFactoryByKey 注册balance工厂
func (dm *driverRegister) RegisterFactoryByKey(key string, factory IBalanceFactory) {
	dm.register.Register(key, factory, true)
	dm.keys = append(dm.keys, key)
}

//Keys 返回所有已注册的key
func (dm *driverRegister) Keys() []string {
	return dm.keys
}

//Register 注册balance工厂到默认balanceFactory注册器
func Register(key string, factory IBalanceFactory) {
	defaultBalanceFactoryRegister.RegisterFactoryByKey(key, factory)
}

//Get 从默认balanceFactory注册器中获取balance工厂
func Get(key string) (IBalanceFactory, bool) {
	return defaultBalanceFactoryRegister.GetFactoryByKey(key)
}

//Keys 返回默认的balanceFactory注册器中所有已注册的key
func Keys() []string {
	return defaultBalanceFactoryRegister.Keys()
}

//GetFactory 获取指定负载均衡算法工厂，若指定的不存在则返回一个已注册的工厂
func GetFactory(name string) (IBalanceFactory, error) {
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
	return factory, nil
}
