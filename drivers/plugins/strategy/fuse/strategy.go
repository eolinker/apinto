package fuse

import (
	fuse_strategy "github.com/eolinker/apinto/drivers/strategy/fuse-strategy"
	"github.com/eolinker/apinto/resources"
	"github.com/eolinker/eosc"
	eoscContext "github.com/eolinker/eosc/eocontext"
)

type Strategy struct {
	id    string
	name  string
	cache *resources.CacheBuilder
}

func (s *Strategy) DoFilter(ctx eoscContext.EoContext, next eoscContext.IChain) (err error) {
	return fuse_strategy.DoStrategy(ctx, next, s.cache.GET())
}

func (s *Strategy) Destroy() {
	return
}

func (s *Strategy) Id() string {
	return s.id
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
