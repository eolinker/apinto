package model_mapping

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
	lock sync.Mutex
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

	// 1. 如果已标记禁用（Stop），从全局管理器中清理并返回
	if cfg.Stop {
		defaultManager.Del(e.Id())
		return nil
	}

	// 2. 创建 Handler
	handler, err := NewHandler(cfg)
	if err != nil {
		return err
	}

	// 3. 注册到全局单例管理器中
	defaultManager.Set(e.Id(), handler, cfg.Filters)
	return nil
}

func (e *executor) Stop() error {
	e.lock.Lock()
	defer e.lock.Unlock()

	defaultManager.Del(e.Id())
	return nil
}

func (e *executor) Destroy() error {
	e.lock.Lock()
	defer e.lock.Unlock()

	defaultManager.Del(e.Id())
	if controller != nil {
		controller.Del(e.Id())
	}
	return nil
}

func (e *executor) CheckSkill(skill string) bool {
	return false
}
