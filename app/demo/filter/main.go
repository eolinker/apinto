package main

import (
	"github.com/eolinker/eosc/http"
	"github.com/eolinker/goku/filter"
)

func main() {
	fi := []http.IFilter{new(TestFilter)}
	filter.NewChain(fi)
}

type TestFilter struct {
}

func (t *TestFilter) DoFilter(ctx http.IHttpContext, next http.IChain) (err error) {
	panic("implement me")
}
