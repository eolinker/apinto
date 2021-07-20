package router_http

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
)

type HeaderSort []string
type HeaderChecker string

func (l HeaderSort) Len() int {
	return len(l)
}

func (l HeaderSort) Less(i, j int) bool {
	// 若l[i]或l[j]中有个是 * ，且只会有一个*结点， * 排到最后
	if l[i] == "*" {
		return false
	}
	if l[j] == "*"{
		return true
	}

	mapI := make(map[string]string)
	mapJ := make(map[string]string)

	_ = json.Unmarshal([]byte(l[i]), &mapI)
	_ = json.Unmarshal([]byte(l[j]), &mapJ)

	// 需要满足key数量多的优先
	if len(mapI) == len(mapJ) {
		// key数量相同则按字母排序从小到大排序，先匹配完的优先
		length := len(mapI)

		KeyArrI := make([]string, 0, length)
		KeyArrJ := make([]string, 0, length)

		for key := range mapI {
			KeyArrI = append(KeyArrI, strings.ToLower(key))
		}
		for key := range mapJ {
			KeyArrJ = append(KeyArrJ, strings.ToLower(key))
		}

		sort.Strings(KeyArrI)
		sort.Strings(KeyArrJ)

		return KeyArrI[length - 1] < KeyArrI[length - 1]

	}
	return len(mapI) > len(mapJ)
}

func (l HeaderSort) Swap(i, j int) {
	l[i],l[j] = l[j],l[i]
}

func(hc HeaderChecker) check(request *http.Request) bool{
	if hc == "*"{
		return true
	}
	headerMap := make(map[string]string)
	json.Unmarshal([]byte(hc), &headerMap)

	for key, value := range headerMap {
		if request.Header.Get(strings.ToLower(key)) != value{
			return false
		}
	}

	return true
}