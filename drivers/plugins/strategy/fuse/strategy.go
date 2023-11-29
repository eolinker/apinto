package fuse

import (
	"sync"

	"github.com/eolinker/eosc"
	eoscContext "github.com/eolinker/eosc/eocontext"

	"github.com/eolinker/apinto/drivers"
	fuse_strategy "github.com/eolinker/apinto/drivers/strategy/fuse-strategy"
	"github.com/eolinker/apinto/resources"
	scope_manager "github.com/eolinker/apinto/scope-manager"
)

type Strategy struct {
	drivers.WorkerBase
	cache        scope_manager.IProxyOutput[resources.ICache]
	redisID      string
	doFilterOnce sync.Once
}

func (s *Strategy) DoFilter(ctx eoscContext.EoContext, next eoscContext.IChain) (err error) {
	s.doFilterOnce.Do(func() {
		s.cache = scope_manager.Auto[resources.ICache](s.redisID, "redis")
	})
	cl := s.cache.List()
	if len(cl) > 0 {
		return fuse_strategy.DoStrategy(ctx, next, cl[0])
	} else {
		return fuse_strategy.DoStrategy(ctx, next, resources.LocalCache())
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
