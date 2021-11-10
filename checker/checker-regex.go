package checker

import (
	"fmt"
	"regexp"
	"strings"

	http_service "github.com/eolinker/eosc/http-service"
)

//checkerAll 实现了Checker接口，能进行正则匹配
type checkerRegexp struct {
	pattern string
	rex     *regexp.Regexp
	tp      http_service.CheckType
}

//Key 返回路由指标检查器带有完整规则符号的检测值
func (t *checkerRegexp) Key() string {
	if t.tp == http_service.CheckTypeRegularG {
		return fmt.Sprintf("~*= %s", t.pattern)
	}
	return fmt.Sprintf("~= %s", t.pattern)

}

//newCheckerAll 创建一个区分大小写的正则匹配类型的检查器
func newCheckerRegexp(pattern string) (*checkerRegexp, error) {
	pattern = fmt.Sprintf("%s", formatPattern(pattern))
	rex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return &checkerRegexp{
		pattern: pattern,
		rex:     rex,
		tp:      http_service.CheckTypeRegular,
	}, nil
}

//newCheckerAll 创建一个不区分大小写的正则匹配类型的检查器
func newCheckerRegexpG(pattern string) (*checkerRegexp, error) {
	pattern = fmt.Sprintf(`(?i)(%s)`, formatPattern(pattern))

	//rex,err:= regexp.CompilePOSIX(pattern)
	//if err!= nil{
	//	return nil,err
	//}
	rex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return &checkerRegexp{
		pattern: pattern,
		rex:     rex,
		tp:      http_service.CheckTypeRegularG,
	}, nil
}

//Value 返回路由指标检查器的检测值
func (t *checkerRegexp) Value() string {
	return t.pattern
}

//Check 判断待检测的路由指标值是否满足检查器的匹配规则
func (t *checkerRegexp) Check(v string, has bool) bool {
	//当待检测的路由指标值满足检查器的正则表达式值时匹配成功
	if !has {
		return false
	}

	return t.rex.MatchString(v)
}

//CheckType 返回检查器的类型值
func (t *checkerRegexp) CheckType() http_service.CheckType {
	return t.tp
}

//formatPattern 格式化正则表达式的值
func formatPattern(pattern string) string {
	pattern = strings.TrimSpace(pattern)
	if len(pattern) == 0 {
		return pattern
	}
	//if strings.HasPrefix(pattern,"/") && strings.HasSuffix(pattern,"/"){
	//	return strings.TrimSuffix(pattern[1:],"/")
	//}
	//if strings.HasPrefix(pattern,"/") && strings.HasSuffix(pattern,"/i"){
	//	return strings.TrimSuffix(pattern[1:],"/i")
	//}

	return pattern

}
