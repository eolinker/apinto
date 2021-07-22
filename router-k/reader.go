package router

import "net/http"

type IReader interface {
	Reader(req *http.Request) (string, bool)
}
