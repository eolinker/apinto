package visit_strategy

import (
	"github.com/eolinker/apinto/strategy"
)

type visitHandler struct {
	name     string
	filter   strategy.IFilter
	priority int
	stop     bool
	rule     ruleHandler
}

type ruleHandler struct {
	visitRule    VisitRule
	visit        bool
	isContinue   bool
	effectFilter strategy.IFilter
}

func newVisitHandler(conf *Config) (*visitHandler, error) {
	filter, err := strategy.ParseFilter(conf.Filters)
	if err != nil {
		return nil, err
	}

	effectFilter, err := strategy.ParseFilter(conf.Rule.InfluenceSphere)
	if err != nil {
		return nil, err
	}

	rule := ruleHandler{
		visitRule:    conf.Rule.VisitRule,
		visit:        conf.Rule.VisitRule == VisitRuleAllow,
		isContinue:   conf.Rule.Continue,
		effectFilter: effectFilter,
	}

	return &visitHandler{
		name:     conf.Name,
		filter:   filter,
		priority: conf.Priority,
		stop:     conf.Stop,
		rule:     rule,
	}, nil
}
