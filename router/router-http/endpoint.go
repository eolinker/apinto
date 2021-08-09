/*
 * Copyright (c) 2021. Lorem ipsum dolor sit amet, consectetur adipiscing elit.
 * Morbi non lorem porttitor neque feugiat blandit. Ut vitae ipsum eget quam lacinia accumsan.
 * Etiam sed turpis ac ipsum condimentum fringilla. Maecenas magna.
 * Proin dapibus sapien vel ante. Aliquam erat volutpat. Pellentesque sagittis ligula eget metus.
 * Vestibulum commodo. Ut rhoncus gravida arcu.
 */

package router_http

import (
	"sync"

	"github.com/eolinker/goku-eosc/router"
	"github.com/eolinker/goku-eosc/router/checker"
	"github.com/eolinker/goku-eosc/service"
)

var _ service.IRouterEndpoint = (*EndPoint)(nil)

type EndPoint struct {
	endpoint router.IEndPoint

	headers []string
	queries []string
	once    sync.Once
}

func (e *EndPoint) Header(name string) (checker.Checker, bool) {
	return e.endpoint.Get(toHeader(name))
}

func (e *EndPoint) Query(name string) (checker.Checker, bool) {
	return e.endpoint.Get(toQuery(name))
}

func (e *EndPoint) initCMD() {
	e.once.Do(func() {
		cs := e.endpoint.CMDs()
		e.headers = make([]string, 0, len(cs))
		e.queries = make([]string, 0, len(cs))
		for _, c := range cs {
			if h, yes := headerName(c); yes {
				e.headers = append(e.headers, h)
				continue
			}
			if q, yes := queryName(c); yes {
				e.queries = append(e.queries, q)
			}
		}
	})

}

func (e *EndPoint) Headers() []string {
	e.initCMD()
	return e.headers
}

func (e *EndPoint) Queries() []string {
	e.initCMD()
	return e.queries
}

func NewEndPoint(endpoint router.IEndPoint) *EndPoint {
	return &EndPoint{endpoint: endpoint}
}

func (e *EndPoint) Location() (checker.Checker, bool) {
	return e.endpoint.Get(cmdLocation)
}
