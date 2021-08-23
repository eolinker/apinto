package router_http

import (
	"github.com/eolinker/goku/router"
	"github.com/eolinker/goku/router/checker"
	"github.com/eolinker/goku/service"
)

//parse 根据路由配置列表生成路由树，且将路由树封装进匹配器并返回
func parse(cs []*Config) (IMatcher, error) {

	count := 0
	for i := range cs {
		hSize := len(cs[i].Hosts)
		mSize := len(cs[i].Methods)

		count += len(cs[i].Rules) * hSize * mSize
	}

	rules := make([]router.Rule, 0, count)

	targets := make(map[string]service.IService)

	//每个路由配置会拆分出多条路由路径，但同一路由配置下的路径都指向同一个服务
	for _, c := range cs {

		//根据路由指标hosts、methods生成相应检查器并封装成RulePath
		hosts := make([]router.RulePath, 0, len(c.Hosts))
		for _, h := range c.Hosts {
			hck, e := checker.Parse(h)
			if e != nil {
				return nil, e
			}
			hosts = append(hosts, router.RulePath{
				CMD:     toHost(),
				Checker: hck,
			})
		}
		methods := make([]router.RulePath, 0, len(c.Methods))
		for _, m := range c.Methods {
			mck, e := checker.Parse(m)
			if e != nil {
				return nil, e
			}
			methods = append(methods, router.RulePath{
				CMD:     toMethod(),
				Checker: mck,
			})
		}
		targets[c.Id] = c.Target

		//若配置里的rules为空时
		if len(c.Rules) == 0 {
			for _, hp := range hosts {
				for _, mp := range methods {
					// 装配路由路径
					pathWithHost := append(make([]router.RulePath, 0, 2), hp, mp)
					rules = append(rules, router.Rule{
						Path:   pathWithHost,
						Target: c.Id,
					})
				}
			}
		}

		for _, r := range c.Rules {

			path, err := r.toPath()
			if err != nil {
				return nil, err
			}
			for _, hp := range hosts {
				for _, mp := range methods {
					pathWithHost := append(make([]router.RulePath, 0, len(path)+2), hp, mp)
					pathWithHost = append(pathWithHost, path...)
					rules = append(rules, router.Rule{
						Path:   pathWithHost,
						Target: c.Id,
					})
				}
			}
		}
	}

	//将所有路由路径组装成一颗路由树
	r, err := router.ParseRouter(rules, NewHttpRouterHelper())
	if err != nil {
		return nil, err
	}

	return &Matcher{
		r:        r,
		services: targets,
	}, nil
}
