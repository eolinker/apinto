package dynamic_params

import (
	"errors"
	"fmt"

	"github.com/eolinker/eosc/log"

	"github.com/eolinker/eosc"
)

var (
	ErrorInvalidBalance    = errors.New("invalid balance")
	defaultFactoryRegister = newFactoryManager()
)

// IFactoryRegister 实现了负载均衡算法工厂管理器
type IFactoryRegister interface {
	RegisterFactoryByKey(key string, factory IDynamicFactory)
	GetFactoryByKey(key string) (IDynamicFactory, bool)
	Keys() []string
}

// driverRegister 实现了IBalanceFactoryRegister接口
type driverRegister struct {
	register eosc.IRegister[IDynamicFactory]
	keys     []string
}

// newFactoryManager 创建负载均衡算法工厂管理器
func newFactoryManager() IFactoryRegister {
	return &driverRegister{
		register: eosc.NewRegister[IDynamicFactory](),
		keys:     make([]string, 0, 10),
	}
}

// GetFactoryByKey 获取指定balance工厂
func (dm *driverRegister) GetFactoryByKey(key string) (IDynamicFactory, bool) {
	o, has := dm.register.Get(key)
	if has {
		log.Debug("GetFactoryByKey:", key, ":has")
		f, ok := o.(IDynamicFactory)
		return f, ok
	}
	return nil, false
}

// RegisterFactoryByKey 注册balance工厂
func (dm *driverRegister) RegisterFactoryByKey(key string, factory IDynamicFactory) {
	err := dm.register.Register(key, factory, true)
	if err != nil {
		log.Debug("RegisterFactoryByKey:", key, ":", err)
		return
	}
	dm.keys = append(dm.keys, key)
}

// Keys 返回所有已注册的key
func (dm *driverRegister) Keys() []string {
	return dm.keys
}

// Register 注册balance工厂到默认balanceFactory注册器
func Register(key string, factory IDynamicFactory) {

	defaultFactoryRegister.RegisterFactoryByKey(key, factory)
}

// Get 从默认balanceFactory注册器中获取balance工厂
func Get(key string) (IDynamicFactory, bool) {
	return defaultFactoryRegister.GetFactoryByKey(key)
}

// Keys 返回默认的balanceFactory注册器中所有已注册的key
func Keys() []string {
	return defaultFactoryRegister.Keys()
}

// GetFactory 获取指定负载均衡算法工厂，若指定的不存在则返回一个已注册的工厂
func GetFactory(name string) (IDynamicFactory, error) {
	factory, ok := Get(name)
	if !ok {
		for _, key := range Keys() {
			factory, ok = Get(key)
			if ok {
				break
			}
		}
		if factory == nil {
			return nil, fmt.Errorf("%s:%w", name, ErrorInvalidBalance)
		}
	}
	return factory, nil
}
