package resources

import (
	"sync"
)

type VectorBuilder struct {
	target string
	once   sync.Once
	vector IVectors
}

func NewVectorBuilder(target string) *VectorBuilder {

	return &VectorBuilder{target: target}
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
