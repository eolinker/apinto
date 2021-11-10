package main

import (
	http_service "github.com/eolinker/eosc/http-service"
	"github.com/eolinker/goku/filter"
)

func main() {
	fi := []http_service.IFilter{new(TestFilter)}
	filter.NewChain(fi)
}

type TestFilter struct {
}

func (t *TestFilter) DoFilter(ctx http_service.IHttpContext, next http_service.IChain) (err error) {
	panic("implement me")
}
