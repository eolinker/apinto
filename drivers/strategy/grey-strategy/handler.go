package grey_strategy

import (
	"github.com/eolinker/apinto/checker"
	"github.com/eolinker/apinto/strategy"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"sync"
)

type GreyHandler struct {
	name       string
	filter     strategy.IFilter
	priority   int
	stop       bool
	rule       *ruleHandler
	ruleFilter strategy.IFilter
}

type ruleHandler struct {
	selectNodeLock *sync.Mutex
	index          int
	keepSession    bool
	nodes          []eocontext.INode
	distribution   string
	flowRobin      *Robin
	matching       []*matchingHandler
}

type matchingHandler struct {
	Type    string
	name    string
	value   string
	checker checker.Checker
}

type matchingHandlerFilters []*matchingHandler

func (m matchingHandlerFilters) Check(ctx eocontext.EoContext) bool {
	for _, handler := range m {
		if !handler.Check(ctx) {
			return false
		}
	}
	return true
}

func (m *matchingHandler) Check(ctx eocontext.EoContext) bool {
	httpCtx, err := http_service.Assert(ctx)
	if err != nil {
		return false
	}

	value := ""
	request := httpCtx.Request()
	switch m.Type {
	case "header":
		value = request.Header().GetHeader(m.name)
	case "query":
		value = request.URI().GetQuery(m.name)
	case "cookie":
		value = request.Header().GetCookie(m.name)
	default:
		return false
	}

	return m.checker.Check(value, true)
}

type flowHandler struct {
	id     int //1为灰度流量 2为正常流量
	weight int
}

func (f *flowHandler) GetId() uint32 {
	return uint32(f.id)
}

func (f *flowHandler) GetWeight() int {
	return f.weight
}

// ABCABCABCABC 轮询从nodes中拿一个节点信息
func (g *GreyHandler) selectNodes() eocontext.INode {
	if len(g.rule.nodes) == 1 {
		return g.rule.nodes[0]
	}
	g.rule.selectNodeLock.Lock()
	defer g.rule.selectNodeLock.Unlock()

	var node eocontext.INode
	if g.rule.index == len(g.rule.nodes)-1 {
		node = g.rule.nodes[g.rule.index]
		g.rule.index = 0
	} else {
		node = g.rule.nodes[g.rule.index]
		g.rule.index++
	}

	return node
}

func NewGreyHandler(conf *Config) (*GreyHandler, error) {
	filter, err := strategy.ParseFilter(conf.Filters)
	if err != nil {
		return nil, err
	}

	rule := &ruleHandler{
		selectNodeLock: &sync.Mutex{},
		keepSession:    conf.Rule.KeepSession,
		nodes:          conf.Rule.GetNodes(),
		distribution:   conf.Rule.Distribution,
	}

	matchHandlers := make([]*matchingHandler, 0)
	ruleFilter := make(matchingHandlerFilters, 0)
	for _, matching := range conf.Rule.Matching {

		check, err := checker.Parse(matching.Value)
		if err != nil {
			return nil, err
		}

		matchingHandlerVal := &matchingHandler{
			Type:    matching.Type,
			name:    matching.Name,
			value:   matching.Value,
			checker: check,
		}

		matchHandlers = append(matchHandlers, matchingHandlerVal)
		ruleFilter = append(ruleFilter, matchingHandlerVal)
	}
	rule.matching = matchHandlers

	//robin算法所需要的数据
	if conf.Rule.Distribution == percent {

		greyFlow := &flowHandler{
			id:     1,
			weight: conf.Rule.Percent,
		}
		normalFlow := &flowHandler{
			id:     2,
			weight: 10000 - greyFlow.weight,
		}

		//总权重10000
		rule.flowRobin = NewRobin(greyFlow, normalFlow)
	}

	return &GreyHandler{
		name:       conf.Name,
		filter:     filter,
		priority:   conf.Priority,
		stop:       conf.Stop,
		rule:       rule,
		ruleFilter: ruleFilter,
	}, nil
}
