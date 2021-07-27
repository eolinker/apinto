package router_http

import (
	"github.com/eolinker/goku-eosc/router"
	"github.com/eolinker/goku-eosc/router/checker"
	"github.com/eolinker/goku-eosc/service"
)

func parse(cs []*Config) (IMatcher, error) {

	count:=0
	for i:=range cs{
		hSize := len(cs[i].Hosts)
	 	mSize := len(cs[i].Methods)

		count += len(cs[i].Rules)*hSize*mSize
	}



	rules :=make([]router.Rule,0,count)

	targets :=make(map[string]service.IService)

	for _,c:=range cs{

		hosts:=make([]router.RulePath,0,len(c.Hosts))
		for _,h:=range c.Hosts{
			hck,e:= checker.Parse(h)
			if e!= nil{
				return nil,e
			}
			hosts = append(hosts, router.RulePath{
				CMD:     toHost(),
				Checker: hck,
			})
		}
		methods:=make([]router.RulePath,0,len(c.Methods))
		for _,m:=range c.Methods{
			mck,e:= checker.Parse(m)
			if e!= nil{
				return nil,e
			}
			methods = append(methods, router.RulePath{
				CMD:     toMethod(),
				Checker: mck,
			})
		}
		targets[c.Id]=c.Target
		for _,r:=range c.Rules{

			path,err:= r.toPath()
			if  err!= nil{
				return nil,err
			}
			for _,hp:=range hosts{
				for _,mp:=range methods{
					pathWithHost := append(make([]router.RulePath,0,len(path)+2),hp,mp)
					pathWithHost = append(pathWithHost,path...)
					rules = append(rules,router.Rule{
						Path:pathWithHost,
						Target:c.Id,
					} )
				}
			}
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

