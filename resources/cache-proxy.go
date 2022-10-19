package resources

import (
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/common/bean"
	"sync"
)

var (
	workers eosc.IWorkers
)

func init() {
	bean.Autowired(&workers)
}

type CacheBuilder struct {
	target string
	once   sync.Once
	cacher ICache
}

func NewCacheBuilder(target string) *CacheBuilder {

	return &CacheBuilder{target: target}
}
func (p *CacheBuilder) GET() ICache {
	p.once.Do(func() {
		if len(p.target) == 0 {
			p.cacher = LocalCache()
			return
		}
		worker, has := workers.Get(p.target)
		if !has || !worker.CheckSkill(CacheSkill) {
			p.cacher = LocalCache()
			return
		}
		p.cacher = worker.(ICache)
	})
	return p.cacher
}
