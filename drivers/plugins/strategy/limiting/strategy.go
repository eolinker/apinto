package limiting

import (
	"github.com/eolinker/apinto/drivers"
	limiting_strategy "github.com/eolinker/apinto/drivers/strategy/limiting-strategy"
	"github.com/eolinker/apinto/resources"
	scope_manager "github.com/eolinker/apinto/scope-manager"
	"github.com/eolinker/eosc"
	eoscContext "github.com/eolinker/eosc/eocontext"
	"sync"
	"time"
)

type Strategy struct {
	drivers.WorkerBase
	buildProxy scope_manager.IProxyOutput[resources.IVectors]
	scalars    limiting_strategy.Scalars
	once       sync.Once
}

func (s *Strategy) DoFilter(ctx eoscContext.EoContext, next eoscContext.IChain) (err error) {
	s.once.Do(func() {
		var iVectors resources.IVectors
		iVectorsList := s.buildProxy.List()
		if len(iVectorsList) > 0 {
			iVectors = iVectorsList[0]
		} else {
			resources.LocalVector()
		}
		s.scalars = limiting_strategy.Scalars{}

		s.scalars.QuerySecond, _ = iVectors.BuildVector("query", time.Second, time.Second/2)
		s.scalars.QueryMinute, _ = iVectors.BuildVector("query", time.Minute, time.Second*10)
		s.scalars.QueryHour, _ = iVectors.BuildVector("query", time.Hour, time.Minute*10)
		s.scalars.TrafficsSecond, _ = iVectors.BuildVector("traffic", time.Second, time.Second/2)
		s.scalars.TrafficsMinute, _ = iVectors.BuildVector("traffic", time.Minute, time.Second*10)
		s.scalars.TrafficsHour, _ = iVectors.BuildVector("traffic", time.Hour, time.Minute*10)
	})

	return limiting_strategy.DoStrategy(ctx, next, &s.scalars)
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
