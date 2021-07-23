package router

import (
	"fmt"
	"github.com/eolinker/goku-eosc/router"
	"github.com/eolinker/goku-eosc/router/checker"
	"github.com/eolinker/goku-eosc/service"
)

const (
	cmdLocation="LOCATION"
	cmdHeader = "HEADER"
	cmdQuery = "QUERY"
)
func toLocation()string{
	return cmdLocation
}
func toHeader(key string) string {
	return fmt.Sprint(cmdHeader ,":",key)
}
func toQuery(key string) string {
	return fmt.Sprint(cmdQuery ,":",key)

}
type  HeaderItem struct {
	Name string
	Pattern string
}
type  QueryItem struct {
	Name string
	Pattern string
}
type Rule struct {
	Location string
	Header   []HeaderItem
	Query    []QueryItem
}

type Config struct {
	Id     string
	Hosts  []string
	Target service.IService
	Rules  []Rule
}

func (r *Rule) toPath()([]router.RulePath ,error) {

	locationChecker,err:= checker.Parse(r.Location)
	if err!= nil{
		return nil,err
	}
	path:=make([]router.RulePath,0,len(r.Header)+len(r.Query)+1)

	path = append(path, router.RulePath{
		CMD:     toLocation(),
		Checker:locationChecker,
	} )

	for _,h:=range r.Header{
		ck,err:= checker.Parse(h.Pattern)
		if err!= nil{
			return  nil,err
		}
		path = append(path, router.RulePath{
			CMD:     toHeader(h.Name),
			Checker: ck,
		})
	}

	for _,h:=range r.Query{
		ck,err:= checker.Parse(h.Pattern)
		if err!= nil{
			return  nil,err
		}
		path = append(path, router.RulePath{
			CMD:     toQuery(h.Name),
			Checker: ck,
		})
	}
	return path,nil
}

