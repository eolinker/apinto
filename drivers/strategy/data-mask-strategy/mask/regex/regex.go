package regex

import (
	"fmt"
	"regexp"

	"github.com/eolinker/apinto/drivers/strategy/data-mask-strategy/mask"
)

func Register() {
	mask.RegisterMaskFactory(mask.MatchRegex, &factory{})
}

type factory struct {
}

func (r *factory) Create(cfg *mask.Rule, maskFunc mask.MaskFunc) (mask.IMaskDriver, error) {
	return newDriver(cfg, maskFunc)
}

var _ mask.IMaskDriver = (*driver)(nil)

type driver struct {
	value   string
	maskCfg *mask.Mask
	mask.MaskFunc
	regexp *regexp.Regexp
}

func newDriver(cfg *mask.Rule, maskFunc mask.MaskFunc) (*driver, error) {
	return &driver{
		value:    cfg.Match.Value,
		maskCfg:  cfg.Mask,
		MaskFunc: maskFunc,
		regexp:   regexp.MustCompile(cfg.Match.Value),
	}, nil
}

func (k *driver) Exec(body []byte) ([]byte, error) {
	if k.MaskFunc == nil || k.regexp == nil {
		return body, nil
	}
	target := k.regexp.ReplaceAllStringFunc(string(body), k.MaskFunc)
	return []byte(target), nil
}

func (k *driver) String() string {
	return fmt.Sprintf("mask driver: regex,value: %s,detail: %v", k.value, k.maskCfg)
}
