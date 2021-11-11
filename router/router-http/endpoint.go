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

	http_service "github.com/eolinker/eosc/http-service"

	"github.com/eolinker/goku/router"
	"github.com/eolinker/goku/service"
)

var _ service.IRouterEndpoint = (*EndPoint)(nil)

//EndPoint 路由端点结构体
type EndPoint struct {
	endpoint router.IEndPoint

	headers []string
	queries []string
	once    sync.Once
}

//Header 通过header的key返回对应指标值的checker
func (e *EndPoint) Header(name string) (http_service.Checker, bool) {
	return e.endpoint.Get(toHeader(name))
}

//Query 通过query的key返回对应指标值的checker
func (e *EndPoint) Query(name string) (http_service.Checker, bool) {
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

//Headers 返回路由端点内header的指标key列表
func (e *EndPoint) Headers() []string {
	e.initCMD()
	return e.headers
}

//Queries 返回路由端点内query的指标key列表
func (e *EndPoint) Queries() []string {
	e.initCMD()
	return e.queries
}

//NewEndPoint 创建
func NewEndPoint(endpoint router.IEndPoint) *EndPoint {
	return &EndPoint{endpoint: endpoint}
}

//Location 返回location指标的checker
func (e *EndPoint) Location() (http_service.Checker, bool) {
	return e.endpoint.Get(cmdLocation)
}
