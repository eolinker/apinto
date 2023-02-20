package utils

import "regexp"

const (
	//regexUrlPathStr 以/开头,能包含字母、数字、下划线、短横线、点号以及"/"
	regexUrlPathStr = `^\/[a-zA-Z0-9\/_\-\.]*$`
	//objectivesExp 校验0.5:0.1,0.9:0.001的格式
	objectivesExp = `^((0\.[0-9]+)\:(0\.[0-9]+)(\,)?)+$`
)

var (
	regexUrlPath     = regexp.MustCompile(regexUrlPathStr)
	objectivesRegexp = regexp.MustCompile(objectivesExp)
)

func CheckUrlPath(path string) bool {
	return regexUrlPath.MatchString(path)
}

func CheckObjectives(objectives string) bool {
	return objectivesRegexp.MatchString(objectives)
}
