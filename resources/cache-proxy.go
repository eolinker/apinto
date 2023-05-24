package resources

import (
	scope_manager "github.com/eolinker/apinto/scope-manager"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/common/bean"
)

var (
	workers eosc.IWorkers
)

func init() {
	bean.Autowired(&workers)
}

func NewCacheBuilder(target string) scope_manager.IProxyOutput[ICache] {
	if len(target) == 0 {
		return scope_manager.Get[ICache]("redis")
	}
	w, has := workers.Get(target)
	if !has || !w.CheckSkill(CacheSkill) {
		return scope_manager.Get[ICache](target)
	}

	return scope_manager.NewProxy(w.(ICache))

}
