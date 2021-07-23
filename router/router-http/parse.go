package router_http

import (
	"github.com/eolinker/goku-eosc/router"
	"github.com/eolinker/goku-eosc/router/checker"
	"github.com/eolinker/goku-eosc/service"
)

func parse(cs []*Config) (IMatcher, error) {

	count:=0
	for i:=range cs{
		hsize := len(cs[i].Hosts)
		if hsize <1{
			hsize = 1
		}
		count += len(cs[i].Rules)*len(cs[i].Hosts)
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
		targets[c.Id]=c.Target
		for _,r:=range c.Rules{

			path,err:= r.toPath()
			if  err!= nil{
				return nil,err
			}
			if len(hosts) >0{
				for _,hp:=range hosts{
					pathWidthHost := append(make([]router.RulePath,0,len(path)+1),hp)
					pathWidthHost = append(pathWidthHost,path...)
					rules = append(rules,router.Rule{
						Path:path,
						Target:c.Id,
					} )
				}
			}else{
				rules = append(rules, router.Rule{
					Path:path,
					Target:c.Id,
				} )
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

