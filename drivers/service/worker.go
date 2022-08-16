package service

import (
	"errors"
	"github.com/eolinker/apinto/service"
)

var (
	ErrorNeedUpstream = errors.New("need upstream")

	ErrorInvalidDiscovery = errors.New("invalid Discovery")
)
var _ service.IService = (*serviceWorker)(nil)

type serviceWorker struct {
	Service
	id     string
	name   string
	driver string
}

//Id 返回服务实例 worker id
func (s *serviceWorker) Id() string {
	return s.id
}

func (s *serviceWorker) Start() error {
	return nil
}

func (s *serviceWorker) Stop() error {

	return nil
}

//CheckSkill 检查目标能力是否存在
func (s *serviceWorker) CheckSkill(skill string) bool {
	return service.CheckSkill(skill)
}
