package grey_strategy

import (
	"fmt"
	"github.com/eolinker/apinto/checker"
	"github.com/eolinker/apinto/discovery"
	"github.com/eolinker/apinto/strategy"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
)

type GreyHandler struct {
	name     string
	filter   strategy.IFilter
	priority int
	stop     bool
	GreyMatch
	discovery.IApp
}

type GreyMatch interface {
	Match(ctx eocontext.EoContext) bool
}

type greyFlow struct {
	flowRobin *Robin
}

type ruleGreyMatch struct {
	ruleFilter strategy.IFilter
}

// Match 按流量计算
func (r *greyFlow) Match(ctx eocontext.EoContext) bool {
	flow := r.flowRobin.Select()
	return flow.GetId() == 1
}

// Match 高级匹配
func (r *ruleGreyMatch) Match(ctx eocontext.EoContext) bool {
	return r.ruleFilter.Check(ctx)
}

type keepSessionGreyFlow struct {
	GreyMatch
}

// Match 保持会话连接
func (k *keepSessionGreyFlow) Match(ctx eocontext.EoContext) bool {

	httpCtx, err := http_service.Assert(ctx)
	if err != nil {
		log.Error("keepSessionGreyFlow err=%s", err.Error())
		return false
	}

	session := httpCtx.Request().Header().GetCookie("session")
	if len(session) == 0 {
		return k.GreyMatch.Match(ctx)
	}

	cookieKey := fmt.Sprintf(cookieName, session)

	cookie := httpCtx.Request().Header().GetCookie(cookieKey)
	if cookie == grey {
		return true
	} else if cookie == normal {
		return false
	}

	if k.GreyMatch.Match(ctx) {
		httpCtx.Response().Headers().Add("Set-Cookie", fmt.Sprintf("%s=%v", cookieKey, grey))
		return true
	} else {
		httpCtx.Response().Headers().Add("Set-Cookie", fmt.Sprintf("%s=%v", cookieKey, normal))
		return false
	}
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

func NewGreyHandler(conf *Config) (*GreyHandler, error) {
	filter, err := strategy.ParseFilter(conf.Filters)
	if err != nil {
		return nil, err
	}

	handler := &GreyHandler{
		name:     conf.Name,
		filter:   filter,
		priority: conf.Priority,
		stop:     conf.Stop,
		IApp:     discovery.NewApp(conf.Rule.GetNodes()).Agent(),
	}

	if conf.Rule.Distribution == percent {
		greyFlowHandler := &flowHandler{
			id:     1,
			weight: conf.Rule.Percent,
		}
		normalFlowHandler := &flowHandler{
			id:     2,
			weight: 10000 - greyFlowHandler.weight,
		}

		robin := NewRobin(greyFlowHandler, normalFlowHandler)
		handler.GreyMatch = &greyFlow{flowRobin: robin}
		if conf.Rule.KeepSession {
			//总权重10000
			handler.GreyMatch = &keepSessionGreyFlow{
				GreyMatch: handler.GreyMatch,
			}
		}

	} else {
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

			ruleFilter = append(ruleFilter, matchingHandlerVal)
		}

		handler.GreyMatch = &ruleGreyMatch{ruleFilter: ruleFilter}
	}

	return handler, nil
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
