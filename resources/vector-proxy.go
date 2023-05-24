package resources

import (
	scope_manager "github.com/eolinker/apinto/scope-manager"
	"sync"
)

type VectorBuilder struct {
	target string
	once   sync.Once
	vector IVectors
}

func NewVectorBuilder(target string) scope_manager.IProxyOutput[IVectors] {
	if len(target) == 0 {
		return scope_manager.Get[IVectors]("redis")
	}
	w, has := workers.Get(target)
	if !has || !w.CheckSkill(CacheSkill) {
		return scope_manager.Get[IVectors](target)
	}

	return scope_manager.NewProxy(w.(IVectors))

}
func (p *VectorBuilder) GET() IVectors {
	p.once.Do(func() {
		if len(p.target) == 0 {
			p.vector = LocalVector()
			return
		}
		worker, has := workers.Get(p.target)
		if !has || !worker.CheckSkill(VectorsSkill) {
			p.vector = LocalVector()
			return
		}
		p.vector = worker.(IVectors)
	})
	return p.vector
}
