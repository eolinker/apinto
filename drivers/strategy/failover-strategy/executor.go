package failover_strategy

import (
	"sync"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

var (
	_ eosc.IWorker        = (*executor)(nil)
	_ eosc.IWorkerDestroy = (*executor)(nil)
)

type executor struct {
	drivers.WorkerBase
	lock    sync.Mutex
	dimKeys []string // 当前策略登记过的维度 keys
}

func (e *executor) Start() error {
	return nil
}

func (e *executor) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	cfg, ok := conf.(*Config)
	if !ok {
		return eosc.ErrorConfigIsNil
	}

	e.lock.Lock()
	defer e.lock.Unlock()

	// 1. 先清理当前 worker 旧登记过的所有维度
	e.cleanExtractors()

	handler, err := NewHandler(cfg)
	if err != nil {
		return err
	}

	// 2. 检查是否有具体的维度过滤条件
	hasFilters := false
	var newDimKeys []string
	for key, val := range cfg.Filters {
		if len(val) == 0 {
			continue
		}
		hasFilters = true
		ex, has := GetExtractor(key)
		if !has {
			ex = NewExtractor(key)
			setExtractor(key, ex)
		}
		ex.Set(e.Id(), val, []IHandler{handler})
		newDimKeys = append(newDimKeys, key)
	}

	// 3. 若无具体过滤条件（例如全局通用策略），注册到全局通用维度（""）
	if !hasFilters {
		ex, has := GetExtractor("")
		if !has {
			ex = NewExtractor("")
			setExtractor("", ex)
		}
		ex.Set(e.Id(), nil, []IHandler{handler})
		newDimKeys = append(newDimKeys, "")
	}

	e.dimKeys = newDimKeys
	return nil
}

func (e *executor) Stop() error {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.cleanExtractors()
	return nil
}

func (e *executor) Destroy() error {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.cleanExtractors()
	if controller != nil {
		controller.Del(e.Id())
	}
	return nil
}

func (e *executor) cleanExtractors() {
	for _, dim := range e.dimKeys {
		if ex, has := GetExtractor(dim); has {
			ex.Del(e.Id())
		}
	}
	e.dimKeys = nil
}

func (e *executor) CheckSkill(skill string) bool {
	return false
}
