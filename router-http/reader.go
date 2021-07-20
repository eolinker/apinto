package router_http

import (
	"net/http"
)

const (
	targetLocation = "location"
	targetHost     = "host"
	targetHeader   = "header"
	targetQuery    = "query"
)

type Reader interface {
	read(request *http.Request) string
}

func CreateReader(targetType string) Reader {

	switch targetType {
	case targetLocation:
		return LocationReader(0)
	case targetHost:
		return HostReader(0)
	}

	return nil
}