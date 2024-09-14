package inner

import (
	"fmt"
	"regexp"

	"github.com/eolinker/apinto/drivers/strategy/data-mask-strategy/mask"
)

func Register() {
	mask.RegisterMaskFactory(mask.MatchInner, &factory{})
}

type factory struct {
}

func (i *factory) Create(cfg *mask.Rule, maskFunc mask.MaskFunc) (mask.IMaskDriver, error) {
	return newDriver(cfg, maskFunc)
}

var _ mask.IMaskDriver = (*driver)(nil)

type driver struct {
	value   string
	maskCfg *mask.Mask
	mask.IInnerMask
}

func newDriver(cfg *mask.Rule, maskFunc mask.MaskFunc) (mask.IMaskDriver, error) {
	var innerMask mask.IInnerMask
	switch cfg.Match.Value {
	case mask.MatchInnerValueName:
		innerMask = NewNameMaskDriver(maskFunc)
	case mask.MatchInnerValuePhone:
		innerMask = newPhoneMaskDriver(maskFunc)
	case mask.MatchInnerValueIDCard:
		innerMask = newIDCardMaskDriver(maskFunc)
	case mask.MatchInnerValueBankCard:
		innerMask = newBankCardMaskDriver(maskFunc)
	case mask.MatchInnerValueAmount:
		innerMask = newAmountMaskDriver(maskFunc)
	case mask.MatchInnerValueDate:
		innerMask = newDateMaskDriver(maskFunc)
	default:
		return nil, fmt.Errorf("invalid inner value: %s", cfg.Match.Value)
	}
	return &driver{
		value:      cfg.Match.Value,
		maskCfg:    cfg.Mask,
		IInnerMask: innerMask,
	}, nil
}

func (k *driver) Exec(body []byte) ([]byte, error) {
	if k.IInnerMask == nil {
		return body, nil
	}
	return k.IInnerMask.Exec(body)
}

func (k *driver) String() string {
	return fmt.Sprintf("mask driver: inner,value: %s,detail: %v", k.value, k.maskCfg)
}

var (
	moneyRegex    = regexp.MustCompile(`(-?\d+)(\.\d{1,2})?`)
	dateTimeRegex = regexp.MustCompile(`(\d{4})[-/.](0[1-9]|1[0-2])[-/.](0[1-9]|[12][0-9]|3[01])(?:[\sT](\d{2}):([0-5][0-9])(:([0-5][0-9]))?)?`)
)

func newAmountMaskDriver(maskFunc mask.MaskFunc) mask.IInnerMask {
	return newCommonMaskDriver(maskFunc, []*regexp.Regexp{moneyRegex})
}

func newDateMaskDriver(maskFunc mask.MaskFunc) mask.IInnerMask {
	return newCommonMaskDriver(maskFunc, []*regexp.Regexp{dateTimeRegex})
}

type commonMaskDriver struct {
	mask.MaskFunc
	regexps []*regexp.Regexp
}

func newCommonMaskDriver(maskFunc mask.MaskFunc, regexps []*regexp.Regexp) *commonMaskDriver {
	return &commonMaskDriver{MaskFunc: maskFunc, regexps: regexps}
}

func (i *commonMaskDriver) Exec(body []byte) ([]byte, error) {
	if i.MaskFunc == nil || i.regexps == nil {
		return body, nil
	}
	for _, re := range i.regexps {

		body = re.ReplaceAllFunc(body, func(bytes []byte) []byte {
			return []byte(i.MaskFunc(string(bytes)))
		})
	}
	return body, nil
}

// 定义递归函数来遍历并修改包含 `key` 字段的部分
func traverseAndModify(data interface{}, keys []string, f mask.MaskFunc) {
	// 根据不同类型处理
	switch value := data.(type) {
	case map[string]interface{}:
		// 如果是 map，检查是否包含指定字段
		for _, key := range keys {
			if val, exist := value[key]; exist {
				if valStr, ok := val.(string); ok {
					value[key] = f(valStr)
				}
			}
		}

		// 继续遍历嵌套的字段
		for _, v := range value {
			traverseAndModify(v, keys, f)
		}
	case []interface{}:
		// 如果是数组，遍历每个元素
		for _, v := range value {
			traverseAndModify(v, keys, f)
		}
	}
}
