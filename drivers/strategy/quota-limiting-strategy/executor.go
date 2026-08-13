package quota_limiting_strategy

import (
	"context"
	"fmt"
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/resources"
	scope_manager "github.com/eolinker/apinto/scope-manager"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/log"
)

var _ eosc.IWorker = (*executor)(nil)

type executor struct {
	drivers.WorkerBase
	keys string
}

func (e *executor) Start() error {
	return nil
}

func (e *executor) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	cfg, ok := conf.(*Config)
	if !ok {
		return eosc.ErrorConfigType
	}
	e.keys = fmt.Sprintf("%s:%s:*", cfg.PreKey, e.Name())
	addStrategy(e.Id(), cfg)
	return nil
}

func (e *executor) Stop() error {
	
	var cache resources.ICache
	caches := scope_manager.Auto[resources.ICache]("", "redis").List()
	if len(caches) > 0 {
		cache = caches[0]
	}
	if cache == nil {
		cache = resources.LocalCache()
	}
	keys, err := cache.Keys(context.Background(), e.keys).Result()
	if err != nil {
		return err
	}
	count, err := cache.Del(context.Background(), keys...).Result()
	if err != nil {
		return err
	}
	log.Info("quota limiting strategy executor stop, delete keys count: ", count)
	removeStrategy(e.Id())
	
	return nil
}

func (e *executor) CheckSkill(skill string) bool {
	return false
}
