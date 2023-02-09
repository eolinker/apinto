package dubbo_router

import (
	"github.com/eolinker/apinto/utils"
	dubbo_context "github.com/eolinker/eosc/eocontext/dubbo-context"
	"sort"
	"strings"

	"github.com/eolinker/apinto/checker"
	"github.com/eolinker/apinto/router"
)

type RuleType = string

const (
	HttpHeader RuleType = "header"
)

func Parse(rules []router.AppendRule) router.MatcherChecker {
	if len(rules) == 0 {
		return &router.EmptyChecker{}
	}
	rls := make(router.RuleCheckers, 0, len(rules))

	for _, r := range rules {
		ck, _ := checker.Parse(r.Pattern)

		switch strings.ToLower(r.Type) {
		case HttpHeader:
			rls = append(rls, &HeaderChecker{
				name:    r.Name,
				Checker: ck,
			})
		}
	}
	sort.Sort(rls)
	return rls
}

type HeaderChecker struct {
	name string
	checker.Checker
}

func (h *HeaderChecker) Weight() int {
	return int(checker.CheckTypeAll-h.Checker.CheckType()) * len(h.Checker.Value())
}

func (h *HeaderChecker) MatchCheck(req interface{}) bool {
	request, ok := req.(dubbo_context.IRequestReader)
	if !ok {
		return false
	}
	v := utils.InterfaceToString(request.Attachments()[h.name])
	has := len(v) > 0
	return h.Checker.Check(v, has)
}
