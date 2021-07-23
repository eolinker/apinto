package checker

import (
	"fmt"
	"regexp"
	"strings"
)

type checkerRegexp struct {
	pattern string
	rex *regexp.Regexp
	tp CheckType
}

func (t *checkerRegexp) Key() string {
	if t.tp == CheckTypeRegularG{
		return fmt.Sprintf("~*= %s",t.pattern)
	}
	return fmt.Sprintf("~= %s",t.pattern)

}

func newCheckerRegexp(pattern string) (*checkerRegexp,error) {
	pattern =  fmt.Sprintf("/%s/",formatPattern(pattern))
	rex,err:= regexp.Compile(pattern)
	if err!= nil{
		return nil,err
	}
	return &checkerRegexp{
		pattern:pattern,
		rex:rex,
		tp:CheckTypeRegular,
	},nil
}
func newCheckerRegexpG(pattern string,) (*checkerRegexp,error) {
	pattern = fmt.Sprintf("/%s/i",formatPattern(pattern))
	rex,err:= regexp.Compile(pattern)
	if err!= nil{
		return nil,err
	}
	return &checkerRegexp{
		pattern:pattern,
		rex:rex,
		tp:CheckTypeRegularG,
	},nil
}
func (t *checkerRegexp) Value() string {
	return t.pattern
}

func (t *checkerRegexp) Check(v string, has bool) bool {
	if !has{
		return false
	}

	return t.rex.MatchString(v)
}

func (t *checkerRegexp) CheckType() CheckType {
	return t.tp
}

func formatPattern(pattern string)string  {
	pattern = strings.TrimSpace(pattern)
	if len(pattern) ==0{
		return pattern
	}
	if strings.HasPrefix(pattern,"/") && strings.HasSuffix(pattern,"/"){
		return strings.TrimSuffix(pattern[1:],"/")
	}
	if strings.HasPrefix(pattern,"/") && strings.HasSuffix(pattern,"/i"){
		return strings.TrimSuffix(pattern[1:],"/i")
	}

	return pattern

}