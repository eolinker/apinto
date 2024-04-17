package matcher

import http_service "github.com/eolinker/eosc/eocontext/http-context"

func NewStatusCodeMatcher(codes []int) IMatcher {
	return &statusCodeMatcher{codes: codes}
}

type statusCodeMatcher struct {
	codes []int
}

func (m *statusCodeMatcher) Match(ctx http_service.IHttpContext) bool {
	if len(m.codes) < 1 {
		return true
	}
	code := ctx.Response().StatusCode()
	for _, c := range m.codes {
		if c == code {
			return true
		}
	}
	return false
}
