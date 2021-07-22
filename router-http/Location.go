package router_http

import (
	"regexp"
	"strings"
)

const (
	// 正则表达式（区分大小写）
	locationRegularMatchCase = "~"
	// 正则表达式（不区分大小写）
	locationRegularNotMatchCase = "~*"
	// 完全匹配
	locationPerfectMatch = "="
	// 最长匹配前缀
	locationLongestMatch = ""
	// 匹配任意
	locationMatchAny = "*"
)
const (
	// 完全匹配
	locationPerfectMatchIndex = iota
	// 最长匹配前缀
	locationLongestMatchIndex
	// 正则表达式（区分大小写）
	locationRegularMatchCaseIndex
	// 正则表达式（不区分大小写）
	locationRegularNotMatchCaseIndex
	// 匹配任意
	locationMatchAnyIndex
)

type LocationSort []string

func (l LocationSort) Len() int {
	return len(l)
}

func (l LocationSort) Less(i, j int) bool {
	s1,l1 := getLocation(l[i])
	s2,l2:= getLocation(l[j])
	if s1 == s2 { // 优先级
		if len(l1) == len(l2){
			// 长度相同，按字符排
			return l1 < l2
		}
		// 按长度降序
		return l1 > l2
	}
	return s1 < s2
}

func (l LocationSort) Swap(i, j int) {
	l[i],l[j] = l[j],l[i]
}

func getLocation(s string) (int,string)  {

	index := strings.IndexAny(s, "/")
	if index != -1 {
		switch s[:index] {
		case locationPerfectMatch:
			{
				return locationPerfectMatchIndex, s[index:]
			}
		case locationLongestMatch:
			{
				return locationLongestMatchIndex, s[index:]
			}
		case locationRegularMatchCase:
			{
				return locationRegularMatchCaseIndex, s[index:]
			}
		case locationRegularNotMatchCase:
			{
				return locationRegularNotMatchCaseIndex, s[index:]
			}
		case locationMatchAny:
			{
				return locationMatchAnyIndex, "/"
			}
		default:
			return locationLongestMatchIndex, "/"+s
		}
	}

	return locationLongestMatchIndex, "/"+s

}



func createLocation(location string) Checker_One {
	t, s := getLocation(location)
	switch t {
	case locationPerfectMatchIndex:
		return locationPerfectMatchType(s)
	case locationLongestMatchIndex:
		return locationLongestMatchType(s)
	case locationRegularMatchCaseIndex:
		compile :=regexp.MustCompile(s)
		return (*LocationRegularMatchType)(compile)
	case locationRegularNotMatchCaseIndex:
		compile := regexp.MustCompile("(?i)" + strings.Replace(s, " ", "[ \\._-]", -1))
		return (*LocationRegularMatchType)(compile)
	case locationMatchAnyIndex:
		return locationMatchAnyType(s)
	}
	return locationLongestMatchType(s)
}

type locationPerfectMatchType string

func (l locationPerfectMatchType) check(v string) bool {
	return string(l) == v
}

type locationLongestMatchType string

func (l locationLongestMatchType) check(v string) bool {
	return strings.HasPrefix(v, string(l))
}

type LocationRegularMatchType regexp.Regexp

func (l *LocationRegularMatchType) check(v string) bool {
	return (*regexp.Regexp)(l).MatchString(v)
}


type locationMatchAnyType string

func (l locationMatchAnyType) check(v string) bool {
	return true
}
