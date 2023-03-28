package grey_strategy

import (
	"github.com/eolinker/apinto/checker"
	"github.com/eolinker/apinto/discovery"
	"github.com/eolinker/apinto/drivers/discovery/static"
	"github.com/eolinker/apinto/strategy"
	"github.com/eolinker/apinto/upstream/balance"
	"github.com/eolinker/eosc/eocontext"
	"strings"
)

var (
	discoveryAnonymous = static.CreateAnonymous(&static.Config{
		HealthOn: false,
		Health:   nil,
	})
)
var (
	_ IGreyHandler = (*GreyHandler)(nil)
)

type GreyMatch interface {
	Match(ctx eocontext.EoContext) bool
}
type IGreyHandler interface {
	strategy.IFilter
	GreyMatch
	DoGrey(ctx eocontext.EoContext)
	IsStop() bool
	Priority() int
}
type GreyHandler struct {
	name string
	strategy.IFilter
	priority int
	stop     bool
	GreyMatch
	app            discovery.IApp
	balanceHandler eocontext.BalanceHandler
}

func (g *GreyHandler) IsStop() bool {
	return g.stop
}

func (g *GreyHandler) Priority() int {
	return g.priority
}

func (g *GreyHandler) DoGrey(ctx eocontext.EoContext) {
	ctx.SetApp(NewGreyApp(ctx.GetApp(), g.app))
	ctx.SetBalance(g.balanceHandler)
}

func (g *GreyHandler) Close() {
	if g.app != nil {
		g.app.Close()
		g.app = nil
	}
	g.balanceHandler = nil
}

func NewGreyHandler(conf *Config) (*GreyHandler, error) {
	filter, err := strategy.ParseFilter(conf.Filters)
	if err != nil {
		return nil, err
	}

	handler := &GreyHandler{
		name:     conf.Name,
		IFilter:  filter,
		priority: conf.Priority,
		stop:     conf.Stop,
	}
	balanceFactory, err := balance.GetFactory("round-robin")
	if err != nil {
		return nil, err
	}
	balanceHandler, err := balanceFactory.Create()
	if err != nil {
		return nil, err
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
	app, err := discoveryAnonymous.GetApp(strings.Join(conf.Rule.Nodes, ";"))
	if err != nil {
		return nil, err
	}
	handler.app = app
	handler.balanceHandler = balanceHandler
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
