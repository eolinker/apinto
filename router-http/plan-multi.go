package router_http

import (
	"github.com/eolinker/goku-eosc/router"
	"net/http"
)

type Checker_Multi interface {
	check(request *http.Request) bool
}

// 适合从request检测多个值的类型，如header、query，读取request中需要检测的值在check中执行
type _Plan_Multi struct {
	checkers []Checker_Multi
	nexts    []router.IRouterHandler
}

func (p *_Plan_Multi) Match(request *http.Request) (string, bool) {

	for i, c := range p.checkers {
		if c.check(request) {
			return p.nexts[i].Match(request)
		}
	}
	return "", false
}
