package router_http

import "net/http"

type endPoint string

func (e endPoint) Match(request *http.Request) (string, bool) {
	return string(e),true
}