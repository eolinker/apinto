package inner

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/eolinker/apinto/drivers/strategy/data-mask-strategy/mask"
)

var (
	idCardRegx = regexp.MustCompile(`[a-zA-Z0-9]{0,1}[1-9]\d{5}(19|20)\d{2}(0[1-9]|1[0-2])(0[1-9]|[12]\d|3[01])\d{3}[\dXx][a-zA-Z0-9]{0,1}`)
	//idCardRegx       = regexp.MustCompile(`[1-9]\d{5}(19|20)\d{2}(0[1-9]|1[0-2])(0[1-9]|[12]\d|3[01])\d{3}[\dXx]`)
)

// 权重系数
var weight = []int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}

// 校验码对应表
var checkCodeMap = []string{"1", "0", "X", "9", "8", "7", "6", "5", "4", "3", "2"}

// 验证身份证号的最后一位校验码
func validateIDCard(id string) bool {
	// 身份证必须为18位
	if len(id) != 18 {
		return false
	}

	// 前17位必须都是数字
	id17 := id[:17]
	if _, err := strconv.Atoi(id17); err != nil {
		return false
	}

	// 计算加权和
	sum := 0
	for i := 0; i < 17; i++ {
		num, _ := strconv.Atoi(string(id[i]))
		sum += num * weight[i]
	}

	// 求模11的余数
	mod := sum % 11

	// 校验码
	checkCode := checkCodeMap[mod]

	// 比较第18位是否匹配
	return strings.ToUpper(string(id[17])) == checkCode
}

func newIDCardMaskDriver(maskFunc mask.MaskFunc) mask.IInnerMask {
	return newCommonMaskDriver(func(origin string) string {
		if len(origin) != 18 {
			return origin
		}
		// 判断是否是合法的身份证号，若非合法，则不做处理，避免误判
		if !validateIDCard(origin) {
			return origin
		}
		return maskFunc(origin)
	}, []*regexp.Regexp{idCardRegx})
}
