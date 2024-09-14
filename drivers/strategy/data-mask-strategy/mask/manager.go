package mask

import (
	"github.com/eolinker/eosc"
)

var (
	maskManager = NewMaskManager()
)

type IMaskFactory interface {
	Create(cfg *Rule, maskFunc MaskFunc) (IMaskDriver, error)
}

type MaskManager struct {
	factories eosc.Untyped[string, IMaskFactory]
}

func NewMaskManager() *MaskManager {
	return &MaskManager{
		factories: eosc.BuildUntyped[string, IMaskFactory](),
	}
}

func (m *MaskManager) Register(name string, driver IMaskFactory) {
	m.factories.Set(name, driver)
}

func (m *MaskManager) Get(name string) (IMaskFactory, bool) {
	return m.factories.Get(name)
}

func (m *MaskManager) Del(name string) {
	m.factories.Del(name)
}

func RegisterMaskFactory(name string, factory IMaskFactory) {
	maskManager.Register(name, factory)
}

func GetMaskFactory(name string) (IMaskFactory, bool) {
	return maskManager.Get(name)
}
