package service

import (
	"errors"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/service"
)

var (
	ErrorNeedUpstream = errors.New("need upstream")

	ErrorInvalidDiscovery = errors.New("invalid Discovery")
)
var _ service.IService = (*serviceWorker)(nil)

type serviceWorker struct {
	drivers.WorkerBase
	Service
}

func (s *serviceWorker) Start() error {
	return nil
}

func (s *serviceWorker) Stop() error {
	if s.app != nil {
		s.app.Close()
		s.app = nil
	}
	return nil
}

func (s *serviceWorker) Destroy() error {
	if s.app != nil {
		s.app.Close()
		s.app = nil
	}
	return nil
}

// CheckSkill 检查目标能力是否存在
func (s *serviceWorker) CheckSkill(skill string) bool {
	return service.CheckSkill(skill)
}
