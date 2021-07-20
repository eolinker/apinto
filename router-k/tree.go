package router

import "net/http"

type Tree struct {
}

func (t *Tree) Match(req *http.Request) (http.Handler, bool) {
	panic("implement me")
}

func parse(cs []*Config) (IMatcher, error) {
	//todo parse config to tree
	return &Tree{}, nil
}
