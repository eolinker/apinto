package utils

import (
	"net/http"
	"strings"
)

func HeaderToString(h http.Header) string {
	if h == nil {
		return ""
	}
	s := &strings.Builder{}
	h.Write(s)
	return s.String()
}
