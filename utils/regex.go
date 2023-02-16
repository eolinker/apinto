package utils

import "regexp"

const (
	//regexUrlPathStr 以/开头,能包含字母、数字、下划线、短横线、点号以及"/"
	regexUrlPathStr = `^\/[a-zA-Z0-9\/_\-\.]*$`
)

var (
	regexUrlPath = regexp.MustCompile(regexUrlPathStr)
)

func CheckUrlPath(path string) bool {
	return regexUrlPath.MatchString(path)
}
