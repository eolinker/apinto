package inner

import (
	"encoding/json"

	"github.com/eolinker/apinto/drivers/strategy/data-mask-strategy/mask"
)

type nameMaskDriver struct {
	maskFunc mask.MaskFunc
}

func NewNameMaskDriver(maskFunc mask.MaskFunc) mask.IInnerMask {
	return &nameMaskDriver{maskFunc: maskFunc}
}

func (i *nameMaskDriver) Exec(body []byte) ([]byte, error) {
	if i.maskFunc == nil {
		// 未设置脱敏方法，不做处理
		return body, nil
	}
	var jsonData interface{}
	err := json.Unmarshal(body, &jsonData)
	if err == nil {
		traverseAndModify(jsonData, []string{
			"name", "cname",
		}, i.maskFunc)

		return json.Marshal(jsonData)
	}
	return body, nil
}
