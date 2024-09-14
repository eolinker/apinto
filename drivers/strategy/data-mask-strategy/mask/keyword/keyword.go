package keyword

import (
	"fmt"
	"strings"

	"github.com/eolinker/apinto/drivers/strategy/data-mask-strategy/mask"
)

func Register() {
	mask.RegisterMaskFactory(mask.MatchKeyword, &factory{})
}

type factory struct {
}

func (k *factory) Create(cfg *mask.Rule, maskFunc mask.MaskFunc) (mask.IMaskDriver, error) {
	return NewKeywordMaskDriver(cfg, maskFunc)
}

var _ mask.IMaskDriver = (*driver)(nil)

type driver struct {
	value   string
	maskCfg *mask.Mask
	mask.MaskFunc
}

func NewKeywordMaskDriver(cfg *mask.Rule, maskFunc mask.MaskFunc) (*driver, error) {
	return &driver{
		value:    cfg.Match.Value,
		maskCfg:  cfg.Mask,
		MaskFunc: maskFunc,
	}, nil
}

func (k *driver) Exec(body []byte) ([]byte, error) {
	if k.MaskFunc == nil {
		return body, nil
	}
	target := strings.Replace(string(body), k.value, k.MaskFunc(k.value), -1)
	return []byte(target), nil
}

func (k *driver) String() string {
	return fmt.Sprintf("mask driver: keyword,value: %s,detail: %v", k.value, k.maskCfg)
}
