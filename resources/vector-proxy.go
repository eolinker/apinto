package resources

import (
	scope_manager "github.com/eolinker/apinto/scope-manager"
)

func NewVectorBuilder(target string) scope_manager.IProxyOutput[IVectors] {
	if len(target) == 0 {
		return scope_manager.Get[IVectors]("redis")
	}
	w, has := workers.Get(target)
	if !has || !w.CheckSkill(VectorsSkill) {
		return scope_manager.Get[IVectors](target)
	}

	return scope_manager.NewProxy(w.(IVectors))

}
