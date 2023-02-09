package router

import (
	"fmt"
	"sort"
	"strings"

	"github.com/eolinker/apinto/checker"
)

const All = "*"

type RuleType = string

type AppendRule struct {
	Type    string
	Name    string
	Pattern string
}

type AppendRules []AppendRule

func (as AppendRules) Len() int {
	return len(as)
}

func (as AppendRules) Less(i, j int) bool {
	if as[i].Type != as[j].Type {
		return as[i].Type < as[j].Type
	}
	if as[i].Name != as[j].Name {
		return as[i].Name < as[j].Name
	}
	return as[i].Pattern < as[j].Pattern
}

func (as AppendRules) Swap(i, j int) {
	as[i], as[j] = as[j], as[i]
}

func Key(rules []AppendRule) string {
	if len(rules) == 0 {
		return All
	}
	strs := make([]string, 0, len(rules))
	rs := make(AppendRules, len(rules))
	copy(rs, rules)
	sort.Sort(rs)
	for _, r := range rs {
		strs = append(strs, fmt.Sprintf("%s[%s]=%s", strings.ToLower(r.Type), r.Name, strings.TrimSpace(r.Pattern)))
	}

	return strings.Join(strs, "&")
}

type EmptyChecker struct {
}

func (e *EmptyChecker) Weight() int {
	return 0
}

func (e *EmptyChecker) MatchCheck(request interface{}) bool {
	return true
}

type MatcherChecker interface {
	MatchCheck(request interface{}) bool
	Weight() int
}
type MatcherCheckerItem interface {
	checker.Checker
	MatcherChecker
}
type RuleCheckers []MatcherCheckerItem

func (rs RuleCheckers) Weight() int {
	w := len(rs)
	for _, i := range rs {
		w += i.Weight()
	}
	return w
}

func (rs RuleCheckers) MatchCheck(request interface{}) bool {
	for _, mc := range rs {
		if !mc.MatchCheck(request) {
			return false
		}
	}
	return true
}

func (rs RuleCheckers) Len() int {
	return len(rs)
}

func (rs RuleCheckers) Less(i, j int) bool {
	ri, rj := rs[i], rs[j]
	//按匹配规则优先级排序
	if ri.CheckType() != rj.CheckType() {
		return ri.CheckType() < rj.CheckType()
	}

	//按长度排序, 优先级 长>短
	vl := len(ri.Value()) - len(rj.Value())
	if vl != 0 {
		return vl > 0
	}
	return ri.Value() < rj.Value()
}

func (rs RuleCheckers) Swap(i, j int) {
	rs[i], rs[j] = rs[j], rs[i]
}
