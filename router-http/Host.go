package router_http

import "strings"

const (
	// 完整域名匹配
	hostPerfectMatchHostIndex = iota
	// 子域名匹配
	hostLongestMatchHostIndex
	// 任意匹配
	hostMatchAnyHostIndex
)

type HostSort []string

func (l HostSort) Len() int {
	return len(l)
}

func (l HostSort) Less(i, j int) bool {
	s1,l1 := getHost(l[i])
	s2,l2:= getHost(l[j])
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

func (l HostSort) Swap(i, j int) {
	l[i],l[j] = l[j],l[i]
}

func createHost(location string) Checker_One {
	t,s:= getHost(location)
	switch t {
	case hostPerfectMatchHostIndex:
		return hostPerfectMatchType(s)
	case hostLongestMatchHostIndex:
		return hostSubHostMatchType(s)
	default:
		return hostMatchAnyType(s)
	}
}

func getHost(s string) (int,string)  {
	if s == "*"{
		return hostMatchAnyHostIndex, s
	}

	if index := strings.LastIndex(s,"*"); index != -1{
		return hostLongestMatchHostIndex, s[index + 1:]
	}else{
		return hostPerfectMatchHostIndex, s
	}
}

type hostPerfectMatchType string

func (l hostPerfectMatchType) check(v string) bool {
	return string(l) == v
}

type hostSubHostMatchType string

func (l hostSubHostMatchType) check(v string) bool {
	return strings.HasSuffix(v, string(l))
}

type hostMatchAnyType string

func (l hostMatchAnyType) check(v string) bool {
	return true
}

