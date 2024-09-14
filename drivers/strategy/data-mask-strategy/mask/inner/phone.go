package inner

import (
	"regexp"

	"github.com/eolinker/apinto/drivers/strategy/data-mask-strategy/mask"
)

var (
	phoneRegx       = regexp.MustCompile(`1[3-9]\d{9}`)
	fatherPhoneRegx = regexp.MustCompile(`[a-zA-Z0-9]{0,1}\d{11}[a-zA-Z0-9]{0,1}`)
)

func newPhoneMaskDriver(maskFunc mask.MaskFunc) mask.IInnerMask {
	return &phoneMaskDriver{
		MaskFunc: maskFunc,
	}
}

type phoneMaskDriver struct {
	mask.MaskFunc
}

func (i *phoneMaskDriver) Exec(body []byte) ([]byte, error) {
	if i.MaskFunc == nil {
		return body, nil
	}

	body = fatherPhoneRegx.ReplaceAllFunc(body, func(b []byte) []byte {
		if len(b) > 11 || !phoneRegx.Match(b) {
			return b
		}

		return []byte(i.MaskFunc(string(b)))
	})
	return body, nil
}
