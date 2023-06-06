package scope_manager

import (
	"github.com/eolinker/eosc"
	"sync"
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

func Set(name string, value interface{}, scopes ...string) {
	locker.Lock()
	defer locker.Unlock()
	del(name)

	set(name, value, append(scopes, name))
	rebuild()
}

func set(name string, value interface{}, scopes []string) {
	if len(scopes) < 1 {
		return
	}
	connScope.Set(name, scopes)
	for _, scope := range scopes {
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
	scopes, has := connScope.Del(name)
	if has {
		for _, scope := range scopes {
			output, has := connOutput.Get(scope)
			if !has {
				continue
			}
			output.Del(name)
		}
	}
}
