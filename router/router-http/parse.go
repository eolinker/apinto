package router_http

import (
	"github.com/eolinker/goku-eosc/router"
	"github.com/eolinker/goku-eosc/service"
)

func parse(cs []*Config) (IMatcher, error) {

	count:=0
	for i:=range cs{

		count += len(cs[i].Rules)
	}

	rules :=make([]router.Rule,0,count)


	targets :=make(map[string]service.IService)
	for _,c:=range cs{

		targets[c.Id]=c.Target
		for _,r:=range c.Rules{

			path,err:= r.toPath()
			if  err!= nil{
				return nil,err
			}
			rules = append(rules, router.Rule{
				Path:path,
				Target:c.Id,
			} )
		}
	}
	r,err:=router.ParseRouter(rules,NewHttpRouterHelper())
	if err!= nil{
		return nil,err
	}

	return &Matcher{
		r:        r,
		services: targets,
	},nil
}

