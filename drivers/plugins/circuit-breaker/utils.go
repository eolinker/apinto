package circuit_breaker

import (
	http_service "github.com/eolinker/eosc/context/http-context"
	"strconv"
	"strings"
)

// 匹配接口状态码
func MatchStatusCode(matchStatusCodes string, ctx http_service.IHttpContext) bool {

	statusCode := ctx.Response().StatusCode()
	if strings.Contains(matchStatusCodes, strconv.Itoa(statusCode)) {
		return true
	} else {
		return false
	}

}

// 判断是否可写body
func checkAllowBody(statusCode int) bool {
	allow := true
	switch {
	case statusCode >= 100 && statusCode <= 199:
		allow = false
	case statusCode == 204:
		allow = false
	case statusCode == 304:
		allow = false
	}
	return allow
}

func writeResponse(ctx http_service.IHttpContext, headers map[string]string, body string, statusCode int) {
	for key, value := range headers {
		ctx.Response().SetHeader(key, value)
	}
	ctx.Response().SetStatus(statusCode, strconv.Itoa(statusCode))
	if checkAllowBody(statusCode) {
		ctx.Response().SetBody([]byte(body))
	}
}
