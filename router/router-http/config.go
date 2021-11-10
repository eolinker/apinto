package router_http

import (
	http_service "github.com/eolinker/eosc/http-service"
	"github.com/eolinker/goku/router"
	"github.com/eolinker/goku/service"
)

//HeaderItem HeaderItem
type HeaderItem struct {
	Name    string
	Pattern string
}

//QueryItem QueryItem
type QueryItem struct {
	Name    string
	Pattern string
}

//Rule 路由Rule
type Rule struct {
	Location string
	Header   []HeaderItem
	Query    []QueryItem
}

//Cert 证书结构体
type Cert struct {
	Crt string
	Key string
}

//Config http路由实例配置结构体
type Config struct {
	ID   string
	Name string
	//Cert    []Cert
	Hosts   []string
	Methods []string
	Target  service.IService
	Rules   []Rule
}

//toPath 根据路由指标Location、Header、Query生成相应Checker并封装成RulePath切片返回
func (r *Rule) toPath() ([]router.RulePath, error) {

	path := make([]router.RulePath, 0, len(r.Header)+len(r.Query)+1)

	if len(r.Location) > 0 {
		locationChecker, err := http_service.Parse(r.Location)
		if err != nil {
			return nil, err
		}
		path = append(path, router.RulePath{
			CMD:     toLocation(),
			Checker: locationChecker,
		})
	}

	for _, h := range r.Header {
		ck, err := http_service.Parse(h.Pattern)
		if err != nil {
			return nil, err
		}
		path = append(path, router.RulePath{
			CMD:     toHeader(h.Name),
			Checker: ck,
		})
	}

	for _, h := range r.Query {
		ck, err := http_service.Parse(h.Pattern)
		if err != nil {
			return nil, err
		}
		path = append(path, router.RulePath{
			CMD:     toQuery(h.Name),
			Checker: ck,
		})
	}
	return path, nil
}
