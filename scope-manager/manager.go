package scope_manager

import (
	"sync"

	"github.com/eolinker/eosc"
)

var (
	scopes     eosc.Untyped[string, *_Proxy]
	connScope  eosc.Untyped[string, []string]
	connOutput eosc.Untyped[string, eosc.Untyped[string, interface{}]]
	locker     sync.Mutex
)

func init() {
	scopes = eosc.BuildUntyped[string, *_Proxy]()
	connScope = eosc.BuildUntyped[string, []string]()
	connOutput = eosc.BuildUntyped[string, eosc.Untyped[string, interface{}]]()
}

func Get[T any](scopeName string) IProxyOutput[T] {
	proxy, has := scopes.Get(scopeName)
	if !has {
		locker.Lock()
		defer locker.Unlock()
		proxy, has = scopes.Get(scopeName)
		if !has {
			proxy = newProxy()
			scopes.Set(scopeName, proxy)
		}
	}
	return create[T](proxy)
}

func Set(name string, value interface{}, ss ...string) {
	locker.Lock()
	defer locker.Unlock()
	del(name)

	set(name, value, ss)
	rebuild()
}

func set(name string, value interface{}, ss []string) {
	if len(ss) < 1 {
		return
	}
	connScope.Set(name, ss)
	for _, scope := range ss {
		output, has := connOutput.Get(scope)
		if !has {
			output = eosc.BuildUntyped[string, interface{}]()
			connOutput.Set(scope, output)
		}
		output.Set(name, value)
	}
}

func rebuild() {
	outputs := connOutput.All()
	for key, value := range outputs {
		proxy, has := scopes.Get(key)
		if !has {
			proxy = newProxy()
		}
		proxy.Set(value.List())
		scopes.Set(key, proxy)
	}
}

func Del(name string) {
	locker.Lock()
	defer locker.Unlock()
	del(name)
	rebuild()
}

func del(name string) {
	ss, has := connScope.Del(name)
	if has {
		for _, scope := range ss {
			output, has := connOutput.Get(scope)
			if !has {
				continue
			}
			output.Del(name)
		}
	}
}
