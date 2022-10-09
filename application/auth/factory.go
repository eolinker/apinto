package auth

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/eolinker/apinto/application"
	"github.com/eolinker/eosc/log"

	"github.com/eolinker/eosc"
)

var (
	ErrorInvalidAuth                         = errors.New("invalid auth")
	defaultAuthFactoryRegister               = newAuthFactoryManager()
	_                          eosc.ISetting = defaultAuthFactoryRegister
)

//IAuthFactory 鉴权工厂方法
type IAuthFactory interface {
	Create(tokenName string, position string, rule interface{}) (application.IAuth, error)
	Alias() []string
	Render() interface{}
	ConfigType() reflect.Type
	UserType() reflect.Type
}

//IAuthFactoryRegister 实现了鉴权工厂管理器
type IAuthFactoryRegister interface {
	RegisterFactoryByKey(key string, factory IAuthFactory)
	GetFactoryByKey(key string) (IAuthFactory, bool)
	Keys() []string
	Alias() map[string]string
}

//driverRegister 驱动注册器
type driverRegister struct {
	register    eosc.IRegister
	keys        []string
	driverAlias map[string]string
	render      map[string]interface{}
}

func (dm *driverRegister) Check(cfg interface{}) (profession, name, driver, desc string, err error) {
	return
}

func (dm *driverRegister) AllWorkers() []string {
	return nil
}

func (dm *driverRegister) Mode() eosc.SettingMode {
	return eosc.SettingModeReadonly
}

func (dm *driverRegister) ConfigType() reflect.Type {
	return nil
}

func (dm *driverRegister) Set(conf interface{}) (err error) {
	return
}

func (dm *driverRegister) Get() interface{} {
	rs := make([]interface{}, 0, len(dm.keys))
	for _, key := range dm.keys {
		if v, ok := dm.render[key]; ok {
			rs = append(rs, map[string]interface{}{
				"name":   key,
				"render": v,
			})
		}
	}
	return rs
}

func (dm *driverRegister) ReadOnly() bool {
	return true
}

//newAuthFactoryManager 创建auth工厂管理器
func newAuthFactoryManager() *driverRegister {
	return &driverRegister{
		register:    eosc.NewRegister(),
		keys:        make([]string, 0, 10),
		driverAlias: make(map[string]string),
		render:      map[string]interface{}{},
	}
}

//GetFactoryByKey 获取指定auth工厂
func (dm *driverRegister) GetFactoryByKey(key string) (IAuthFactory, bool) {
	o, has := dm.register.Get(key)
	if has {
		f, ok := o.(IAuthFactory)
		return f, ok
	}
	return nil, false
}

//RegisterFactoryByKey 注册auth工厂
func (dm *driverRegister) RegisterFactoryByKey(key string, factory IAuthFactory) {
	err := dm.register.Register(key, factory, true)
	if err != nil {
		log.Debug("RegisterFactoryByKey:", key, ":", err)
		return
	}
	dm.keys = append(dm.keys, key)
	for _, alias := range factory.Alias() {
		dm.driverAlias[strings.ToLower(alias)] = key
		dm.render[key] = factory.Render()
	}
}

//Keys 返回所有已注册的key
func (dm *driverRegister) Keys() []string {
	return dm.keys
}

func (dm *driverRegister) Alias() map[string]string {
	return dm.driverAlias
}

//FactoryRegister 注册auth工厂到默认auth工厂注册器
func FactoryRegister(key string, factory IAuthFactory) {

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

func Alias() map[string]string {
	return defaultAuthFactoryRegister.Alias()
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
