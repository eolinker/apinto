package cache

import (
	"github.com/eolinker/apinto/drivers"
	cache_strategy "github.com/eolinker/apinto/drivers/strategy/cache-strategy"
	"github.com/eolinker/apinto/resources"
	scope_manager "github.com/eolinker/apinto/scope-manager"
	"github.com/eolinker/eosc"
	eoscContext "github.com/eolinker/eosc/eocontext"
)

type Strategy struct {
	drivers.WorkerBase
	cache scope_manager.IProxyOutput[resources.ICache]
}

func (s *Strategy) DoFilter(ctx eoscContext.EoContext, next eoscContext.IChain) (err error) {
	cl := s.cache.List()
	if len(cl) > 0 {
		return cache_strategy.DoStrategy(ctx, next, cl[0])
	} else {
		return cache_strategy.DoStrategy(ctx, next, resources.LocalCache())
	}
}

func (s *Strategy) Destroy() {
	return
}

func (s *Strategy) Start() error {
	return nil
}

func (s *Strategy) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	return nil
}

func (s *Strategy) Stop() error {
	return nil
}

func (s *Strategy) CheckSkill(skill string) bool {
	return eoscContext.FilterSkillName == skill
}
