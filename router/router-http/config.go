package router_http

import (
	"github.com/eolinker/goku/router"
	"github.com/eolinker/goku/router/checker"
	"github.com/eolinker/goku/service"
)

type HeaderItem struct {
	Name    string
	Pattern string
}
type QueryItem struct {
	Name    string
	Pattern string
}
type Rule struct {
	Location string
	Header   []HeaderItem
	Query    []QueryItem
}

type Cert struct {
	Crt string
	Key string
}

type Config struct {
	Id       string
	Name     string
	Protocol string
	Cert     []Cert
	Hosts    []string
	Methods  []string
	Target   service.IService
	Rules    []Rule
}

func (r *Rule) toPath() ([]router.RulePath, error) {

	path := make([]router.RulePath, 0, len(r.Header)+len(r.Query)+1)

	if len(r.Location) > 0 {
		locationChecker, err := checker.Parse(r.Location)
		if err != nil {
			return nil, err
		}
		path = append(path, router.RulePath{
			CMD:     toLocation(),
			Checker: locationChecker,
		})
	}

	for _, h := range r.Header {
		ck, err := checker.Parse(h.Pattern)
		if err != nil {
			return nil, err
		}
		path = append(path, router.RulePath{
			CMD:     toHeader(h.Name),
			Checker: ck,
		})
	}

	for _, h := range r.Query {
		ck, err := checker.Parse(h.Pattern)
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
