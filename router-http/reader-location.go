package router_http

import "net/http"

type LocationReader int

func (l LocationReader) read(request *http.Request) string {
	return request.URL.Path
}
