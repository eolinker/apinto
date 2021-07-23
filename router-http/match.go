package router

import (
	http2 "github.com/eolinker/goku-eosc/match/http"
	"net/http"
)

type IMatcher interface {
	Match(req *http.Request) (http.Handler, bool)
}

type Matcher struct {
	reader  http2.IReader
	checker IChecker
}

func (m *Matcher) Match(req *http.Request) (http.Handler, bool) {
	v, has := m.reader.Reader(req)
	if !has {
		return nil, false
	}

}
