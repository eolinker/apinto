package router_http

import (
	"github.com/eolinker/goku-eosc/router"
	"net/http"
)

type Checker_One interface {
	check(v string) bool
}

// 适合从request检测单个值的类型，如host、location
type _Plan_One struct {
	reader   Reader
	checkers []Checker_One
	nexts    []router.IRouterHandler
}

func (p *_Plan_One) Match(request *http.Request) (string, bool) {
	v := p.reader.read(request)

	for i, c := range p.checkers {
		if c.check(v) {
			res, has := p.nexts[i].Match(request)
			// 防止错失后面可能匹配成功的路径
			if !has {
				continue
			}
			return res, has
		}
	}
	return "", false
}
