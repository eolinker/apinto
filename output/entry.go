package output

import (
	http_service "github.com/eolinker/eosc/http-service"

	"github.com/eolinker/eosc/formatter"
)

var (
	_            formatter.IEntry = (*Entry)(nil)
	proxiesChild                  = "proxies"
)

type Entry struct {
	index int
	ctx   http_service.IHttpContext
}

func NewEntry(ctx http_service.IHttpContext, index int) *Entry {
	return &Entry{ctx: ctx, index: index}
}

func (e *Entry) Read(pattern string) string {
	v, ok := rule.Read(pattern, e.index, e.ctx)
	if !ok {
		return ""
	}
	return v
}

func (e *Entry) Children(child string) []formatter.IEntry {
	switch child {
	case proxiesChild:
		length := len(e.ctx.Proxies())
		entries := make([]formatter.IEntry, length)
		for i := 0; i <= length; i++ {
			entries[length] = NewEntry(e.ctx, i)
		}
		return entries
	default:
		length := len(e.ctx.Proxies())
		entries := make([]formatter.IEntry, length)
		for i := 0; i <= length; i++ {
			entries[length] = NewEntry(e.ctx, i)
		}
		return entries
	}
}
