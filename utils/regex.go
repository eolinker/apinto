package utils

import "regexp"

const (
	//regexUrlPathStr 以/开头,能包含字母、数字、下划线、短横线、点号以及"/"
	regexUrlPathStr = `^\/[a-zA-Z0-9\/_\-\.]*$`
	//objectivesExp 校验0.5:0.1,0.9:0.001的格式
	objectivesExp = `^((0\.[0-9]+)\:(0\.[0-9]+)(\,)?)+$`
	// schemeIPPortExp scheme://IP:PORT
	schemeIPPortExp = `^[a-zA-z]+://((2(5[0-5]|[0-4]\d))|[0-1]?\d{1,2})(\.((2(5[0-5]|[0-4]\d))|[0-1]?\d{1,2})){3}:[0-9]+$`
	// schemeDomainPortExp scheme://域名或者域名:端口
	schemeDomainPortExp = `^[a-zA-z]+://[a-zA-Z0-9][-a-zA-Z0-9]{0,62}(\.[a-zA-Z0-9][-a-zA-Z0-9]{0,62})+\.?(:[0-9]+)?$`
	// domainPortExp IP:PORT或者IP
	ipPortExp = `^((2(5[0-5]|[0-4]\d))|[0-1]?\d{1,2})(\.((2(5[0-5]|[0-4]\d))|[0-1]?\d{1,2})){3}(:[0-9]+)?$`
	// domainPortExp 域名或者域名:端口
	domainPortExp = `^[a-zA-Z0-9][-a-zA-Z0-9]{0,62}(\.[a-zA-Z0-9][-a-zA-Z0-9]{0,62})+\.?(:[0-9]+)?$`
)

var (
	regexUrlPath        = regexp.MustCompile(regexUrlPathStr)
	objectivesRegexp    = regexp.MustCompile(objectivesExp)
	schemeIPPortReg     = regexp.MustCompile(schemeIPPortExp)
	schemeDomainPortReg = regexp.MustCompile(schemeDomainPortExp)
	ipPortReg           = regexp.MustCompile(ipPortExp)
	domainPortReg       = regexp.MustCompile(domainPortExp)
)

func CheckUrlPath(path string) bool {
	return regexUrlPath.MatchString(path)
}

// CheckObjectives 检查prometheus objectives配置 校验0.5:0.1,0.9:0.001的格式
func CheckObjectives(objectives string) bool {
	return objectivesRegexp.MatchString(objectives)
}

// IsMatchSchemeIpPort 判断字符串是否符合scheme://ip:port
func IsMatchSchemeIpPort(s string) bool {
	return schemeIPPortReg.MatchString(s)
}

// IsMatchSchemeDomainPort 判断字符串是否符合 scheme://域名或者域名:port
func IsMatchSchemeDomainPort(s string) bool {
	return schemeDomainPortReg.MatchString(s)
}

// IsMatchIpPort 判断字符串是否符合 ip:port或者ip
func IsMatchIpPort(s string) bool {
	return ipPortReg.MatchString(s)
}

// IsMatchDomainPort 判断字符串是否符合 域名或者域名:port
func IsMatchDomainPort(s string) bool {
	return domainPortReg.MatchString(s)
}
