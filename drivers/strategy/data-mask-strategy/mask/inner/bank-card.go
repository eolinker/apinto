package inner

import (
	"regexp"
	"unicode"

	"github.com/eolinker/apinto/drivers/strategy/data-mask-strategy/mask"
)

var cardRegex = regexp.MustCompile(`[a-zA-Z0-9]{0,1}\d{16,19}[a-zA-Z0-9]{0,1}`)

func validateLuhn(cardNumber string) bool {
	var sum int
	// 反向遍历卡号字符串
	nDigits := len(cardNumber)
	parity := nDigits % 2

	for i, r := range cardNumber {
		if !unicode.IsDigit(r) {
			return false // 如果包含非数字字符，则为无效卡号
		}

		digit := int(r - '0') // 将字符转为数字

		// 根据Luhn算法，当索引与奇偶性不同时，数字需要翻倍
		if i%2 == parity {
			digit *= 2
		}

		// 如果数字大于9，减去9
		if digit > 9 {
			digit -= 9
		}

		sum += digit
	}

	// 卡号有效性取决于sum是否可以被10整除
	return sum%10 == 0
}

func newBankCardMaskDriver(maskFunc mask.MaskFunc) mask.IInnerMask {
	return newCommonMaskDriver(func(origin string) string {
		// 判断是否是合法的银行卡号，若非合法，则不做处理，避免误判
		if !validateLuhn(origin) {
			return origin
		}
		size := len(origin)
		if size > 19 {
			return origin
		}
		arr := []byte{
			origin[0],
			origin[1],
		}
		for _, o := range arr {
			if o < 48 || o > 57 {
				return origin
			}
		}
		return maskFunc(origin)
	}, []*regexp.Regexp{cardRegex})
}
