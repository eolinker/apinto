package cors

import "strings"

var allowHeader = map[string]bool{
	"Accept":              true,
	"Accept-Charset":      true,
	"Accept-Encoding":     true,
	"Accept-Language":     true,
	"Accept-Datetime":     true,
	"Authorization":       true,
	"Cache-Control":       true,
	"Connection":          true,
	"Cookie":              true,
	"Content-Length":      true,
	"Content-Type":        true,
	"Date":                true,
	"Expect":              true,
	"From":                true,
	"Host":                true,
	"If-Match":            true,
	"If-Modified-Since":   true,
	"If-None-Match":       true,
	"If-Range":            true,
	"If-Unmodified-Since": true,
	"Max-Forwards":        true,
	"Origin":              true,
	"Pragma":              true,
	"Proxy-Authorization": true,
	"Range":               true,
	"Referer":             true,
	"TE":                  true,
	"User-Agent":          true,
	"Upgrade":             true,
	"Via":                 true,
	"Warning":             true,
}

type ICheck interface {
	Check(check string, isHeader bool) bool
}
type IHeader interface {
	GetOrigin() string
	GetKey() string
}

type Checker struct {
	allowAll  bool
	origin    string
	headerKey string
	addition  map[string]bool
}

func NewChecker(checks string, headerKey string) *Checker {
	h := &Checker{
		origin:    checks,
		headerKey: headerKey,
		allowAll:  false,
		addition:  make(map[string]bool),
	}
	if strings.EqualFold(checks, "*") || strings.EqualFold(checks, "**") {
		h.allowAll = true
	}
	for _, s := range strings.Split(checks, ",") {
		if strings.EqualFold(s, "*") || strings.EqualFold(s, "**") {
			h.allowAll = true
			h.addition = make(map[string]bool)
			break
		}
		h.addition[s] = true
	}
	return h
}

// Check 检查字段是否允许
func (c *Checker) Check(check string, isHeader bool) bool {
	if c.allowAll {
		return true
	}
	if _, ok := c.addition[check]; ok {
		return true
	}
	if isHeader {
		if _, ok := allowHeader[check]; ok {
			return true
		}
	}
	return false
}

func (c *Checker) GetOrigin() string {
	return c.origin
}

func (c *Checker) GetKey() string {
	return c.headerKey
}
