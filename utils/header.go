package utils

import (
	"net/http"
	"strings"
)

//HeaderToString 将header转成字符串
func HeaderToString(h http.Header) string {
	if h == nil {
		return ""
	}
	s := &strings.Builder{}
	h.Write(s)
	return s.String()
}
