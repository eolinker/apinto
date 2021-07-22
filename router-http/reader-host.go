package router_http

import (
	"net/http"
	"strings"
)

type HostReader int

func (h HostReader) read(request *http.Request) string {
	hosts := strings.Split(request.Host, ":")
	return hosts[0]
}
