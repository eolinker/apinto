package json_path

import (
	"fmt"

	"github.com/eolinker/apinto/drivers/strategy/data-mask-strategy/mask"

	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"
)

func Register() {
	mask.RegisterMaskFactory(mask.MatchJsonPath, &factory{})

}

type factory struct {
}

func (j *factory) Create(cfg *mask.Rule, maskFunc mask.MaskFunc) (mask.IMaskDriver, error) {
	return newDriver(cfg, maskFunc)
}

var _ mask.IMaskDriver = (*driver)(nil)

type driver struct {
	value   string
	maskCfg *mask.Mask
	mask.MaskFunc
	expr jp.Expr
}

func newDriver(cfg *mask.Rule, maskFunc mask.MaskFunc) (mask.IMaskDriver, error) {
	expr, err := jp.ParseString(cfg.Match.Value)
	if err != nil {
		return nil, err
	}
	return &driver{
		value:    cfg.Match.Value,
		maskCfg:  cfg.Mask,
		MaskFunc: maskFunc,
		expr:     expr,
	}, nil
}

func (k *driver) Exec(body []byte) ([]byte, error) {
	if k.MaskFunc == nil || k.expr == nil {
		return body, nil
	}
	n, err := oj.Parse(body)
	if err != nil {
		return nil, err
	}
	if n == nil {
		return nil, fmt.Errorf("parse json failed")
	}

	result := k.expr.Get(n)
	if len(result) > 0 {
		val, ok := result[0].(string)
		if ok {
			k.expr.Set(n, k.MaskFunc(val))
		}
	}

	return oj.Marshal(n)
}

func (k *driver) String() string {
	return fmt.Sprintf("mask driver: json-path,value: %s,detail: %v", k.value, k.maskCfg)
}
