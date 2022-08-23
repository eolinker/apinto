package auth

import (
	"errors"
	"fmt"
	"github.com/eolinker/apinto/application"
	"github.com/eolinker/eosc/log"
	
	"github.com/eolinker/eosc"
)

var (
	ErrorInvalidAuth           = errors.New("invalid auth")
	defaultAuthFactoryRegister = newAuthFactoryManager()
)

//IAuthFactory 鉴权工厂方法
type IAuthFactory interface {
	Create(tokenName string, position string, users []*application.User, rule interface{}) (application.IAuth, error)
}

//IAuthFactoryRegister 实现了鉴权工厂管理器
type IAuthFactoryRegister interface {
	RegisterFactoryByKey(key string, factory IAuthFactory)
	GetFactoryByKey(key string) (IAuthFactory, bool)
	Keys() []string
}

//driverRegister 驱动注册器
type driverRegister struct {
	register eosc.IRegister
	keys     []string
}

//newAuthFactoryManager 创建auth工厂管理器
func newAuthFactoryManager() IAuthFactoryRegister {
	return &driverRegister{
		register: eosc.NewRegister(),
		keys:     make([]string, 0, 10),
	}
}

//GetFactoryByKey 获取指定auth工厂
func (dm *driverRegister) GetFactoryByKey(key string) (IAuthFactory, bool) {
	log.Debug("GetFactoryByKey:", key)
	o, has := dm.register.Get(key)
	if has {
		log.Debug("GetFactoryByKey:", key, ":has")
		f, ok := o.(IAuthFactory)
		return f, ok
	}
	return nil, false
}

//RegisterFactoryByKey 注册auth工厂
func (dm *driverRegister) RegisterFactoryByKey(key string, factory IAuthFactory) {
	err := dm.register.Register(key, factory, true)
	log.Debug("RegisterFactoryByKey:", key)
	
	if err != nil {
		log.Debug("RegisterFactoryByKey:", key, ":", err)
		return
	}
	dm.keys = append(dm.keys, key)
}

//Keys 返回所有已注册的key
func (dm *driverRegister) Keys() []string {
	return dm.keys
}

//Register 注册auth工厂到默认auth工厂注册器
func Register(key string, factory IAuthFactory) {
	
	defaultAuthFactoryRegister.RegisterFactoryByKey(key, factory)
}

//Get 从默认auth工厂注册器中获取auth工厂
func Get(key string) (IAuthFactory, bool) {
	return defaultAuthFactoryRegister.GetFactoryByKey(key)
}

//Keys 返回默认的auth工厂注册器中所有已注册的key
func Keys() []string {
	return defaultAuthFactoryRegister.Keys()
}

//GetFactory 获取指定auth工厂，若指定的不存在则返回一个已注册的工厂
func GetFactory(name string) (IAuthFactory, error) {
	factory, ok := Get(name)
	if !ok {
		for _, key := range Keys() {
			factory, ok = Get(key)
			if ok {
				break
			}
		}
		if factory == nil {
			return nil, fmt.Errorf("%s:%w", name, ErrorInvalidAuth)
		}
	}
	return factory, nil
}
