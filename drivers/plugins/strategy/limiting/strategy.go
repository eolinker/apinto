package limiting

import (
	"sync"
	"time"

	"github.com/eolinker/apinto/drivers"
	limiting_strategy "github.com/eolinker/apinto/drivers/strategy/limiting-strategy"
	"github.com/eolinker/apinto/resources"
	scope_manager "github.com/eolinker/apinto/scope-manager"
	"github.com/eolinker/eosc"
	eoscContext "github.com/eolinker/eosc/eocontext"
)

type Strategy struct {
	drivers.WorkerBase
	buildProxy   scope_manager.IProxyOutput[resources.IVectors]
	localScalars *limiting_strategy.Scalars
	redisScalars *limiting_strategy.Scalars
	lastVectorId string
	redisID      string

	lock sync.RWMutex
}

func (s *Strategy) DoFilter(ctx eoscContext.EoContext, next eoscContext.IChain) (err error) {
	iVectorsList := scope_manager.Auto[resources.IVectors](s.redisID, "redis").List()
	var scalars *limiting_strategy.Scalars
	if len(iVectorsList) > 0 {
		iVectors := iVectorsList[0]
		id := iVectors.(eosc.IWorker).Id()
		s.lock.RLock()
		if s.lastVectorId == id {
			scalars = s.redisScalars
			s.lock.RUnlock()
		} else {
			s.lock.RUnlock()
			s.lock.Lock()
			if s.lastVectorId != id {
				s.lastVectorId = id
				redisScalars := &limiting_strategy.Scalars{}
				redisScalars.QuerySecond, _ = iVectors.BuildVector("query", time.Second, time.Second/2)
				redisScalars.QueryMinute, _ = iVectors.BuildVector("query", time.Minute, time.Second*10)
				redisScalars.QueryHour, _ = iVectors.BuildVector("query", time.Hour, time.Minute*10)
				redisScalars.TrafficsSecond, _ = iVectors.BuildVector("traffic", time.Second, time.Second/2)
				redisScalars.TrafficsMinute, _ = iVectors.BuildVector("traffic", time.Minute, time.Second*10)
				redisScalars.TrafficsHour, _ = iVectors.BuildVector("traffic", time.Hour, time.Minute*10)
				s.redisScalars = redisScalars
			}
			scalars = s.redisScalars
			s.lock.Unlock()
		}

	} else {
		if s.localScalars == nil {
			iVectors := resources.LocalVector()
			s.localScalars = &limiting_strategy.Scalars{}
			s.localScalars.QuerySecond, _ = iVectors.BuildVector("query", time.Second, time.Second/2)
			s.localScalars.QueryMinute, _ = iVectors.BuildVector("query", time.Minute, time.Second*10)
			s.localScalars.QueryHour, _ = iVectors.BuildVector("query", time.Hour, time.Minute*10)
			s.localScalars.TrafficsSecond, _ = iVectors.BuildVector("traffic", time.Second, time.Second/2)
			s.localScalars.TrafficsMinute, _ = iVectors.BuildVector("traffic", time.Minute, time.Second*10)
			s.localScalars.TrafficsHour, _ = iVectors.BuildVector("traffic", time.Hour, time.Minute*10)
		}
		s.lastVectorId = ""
		scalars = s.localScalars
	}

	return limiting_strategy.DoStrategy(ctx, next, scalars)
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
