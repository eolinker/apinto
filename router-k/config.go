package router

import "github.com/eolinker/goku-eosc/service"

type Rule struct {
	Location string
	Header   []string
	Query    []string
}

type Config struct {
	Id     string
	Hosts  []string
	Target service.IService
	Rules  []Rule
}
