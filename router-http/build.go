package router_http

import (
	"fmt"
	"github.com/eolinker/goku-eosc/router"
	"sort"
)

//  路由树路径上经过的检测指标， 检测优先级 host>location>header>query
var RouterPathType = []string{
	"host",
	"location",
	"header",
	"query",
}

type Tree map[string]interface{}

func (t Tree) Append(pathValue []string, target string) error {
	if len(pathValue) < 1 {
		return fmt.Errorf("no path exist")
	}
	pv := pathValue[0]
	if len(pathValue) == 1 {
		// 若target冲突 则返回错误  一条路由不能对应多个target
		if _, has := t[pv]; has {
			return fmt.Errorf("router config conflict")
		}
		t[pv] = target
	} else {
		next, has := t[pv]
		if !has {
			nextM := make(Tree)
			err := nextM.Append(pathValue[1:], target)
			if err != nil {
				return err
			}
			t[pv] = nextM
		} else {
			nextM := next.(Tree)
			err := nextM.Append(pathValue[1:], target)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func createRouter(tree Tree, nodesType []string) router.IRouterHandler {
	if len(nodesType) == 0 {
		return nil
	}

	nodeType := nodesType[0]
	nextNodesType := nodesType[1:]

	switch nodeType {
	case targetLocation:
		pl := &_Plan_One{
			reader:   CreateReader(nodeType),
			checkers: nil,
			nexts:    nil,
		}
		sorts := make(LocationSort, 0, len(tree))
		for k := range tree {
			sorts = append(sorts, k)
		}
		sort.Sort(sorts)
		if len(nextNodesType) == 0 {
			for _, s := range sorts {
				v := tree[s]
				pl.checkers = append(pl.checkers, createLocation(s))
				pl.nexts = append(pl.nexts, endPoint(v.(string)))
			}
		} else {
			for _, s := range sorts {
				v := tree[s]
				pl.checkers = append(pl.checkers, createLocation(s))
				pl.nexts = append(pl.nexts, createRouter(v.(Tree), nextNodesType))
			}
		}

		return pl

	case targetHost:
		pl := &_Plan_One{
			reader:   CreateReader(nodeType),
			checkers: nil,
			nexts:    nil,
		}
		sorts := make(HostSort, 0, len(tree))
		for k := range tree {
			sorts = append(sorts, k)
		}
		sort.Sort(sorts)
		if len(nextNodesType) == 0 {
			for _, s := range sorts {
				v := tree[s]
				pl.checkers = append(pl.checkers, createHost(s))
				pl.nexts = append(pl.nexts, endPoint(v.(string)))
			}
		} else {
			for _, s := range sorts {
				v := tree[s]
				pl.checkers = append(pl.checkers, createHost(s))
				pl.nexts = append(pl.nexts, createRouter(v.(Tree), nextNodesType))
			}
		}

		return pl
	case targetHeader:
		pl := &_Plan_Multi{
			checkers: nil,
			nexts:    nil,
		}
		sorts := make(HeaderSort, 0, len(tree))
		for k := range tree {
			sorts = append(sorts, k)
		}
		sort.Sort(sorts)
		if len(nextNodesType) == 0 {
			for _, s := range sorts {
				v := tree[s]
				pl.checkers = append(pl.checkers, HeaderChecker(s))
				pl.nexts = append(pl.nexts, endPoint(v.(string)))
			}
		} else {
			for _, s := range sorts {
				v := tree[s]
				pl.checkers = append(pl.checkers, HeaderChecker(s))
				pl.nexts = append(pl.nexts, createRouter(v.(Tree), nextNodesType))
			}
		}

		return pl
	default: //targetQuery
		pl := &_Plan_Multi{
			checkers: nil,
			nexts:    nil,
		}
		sorts := make(QuerySort, 0, len(tree))
		for k := range tree {
			sorts = append(sorts, k)
		}
		sort.Sort(sorts)
		if len(nextNodesType) == 0 {
			for _, s := range sorts {

				v := tree[s]
				pl.checkers = append(pl.checkers, QueryChecker(s))
				pl.nexts = append(pl.nexts, endPoint(v.(string)))
			}
		} else {
			for _, s := range sorts {

				v := tree[s]
				pl.checkers = append(pl.checkers, QueryChecker(s))
				pl.nexts = append(pl.nexts, createRouter(v.(Tree), nextNodesType))
			}
		}

		return pl
	}
}
