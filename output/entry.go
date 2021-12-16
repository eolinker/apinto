package output

import (
	"strings"

	http_service "github.com/eolinker/eosc/http-service"

	"github.com/eolinker/eosc/formatter"
)

var (
	_            formatter.IEntry = (*Entry)(nil)
	proxiesChild                  = "proxies"
)

type Entry struct {
	ctx http_service.IHttpContext
}

func NewEntry(ctx http_service.IHttpContext) *Entry {
	return &Entry{ctx: ctx}
}

func (e *Entry) Read(pattern string) string {
	v, ok := rule.Read(pattern, e.ctx)
	if !ok {
		return ""
	}
	return v
}

func (e *Entry) Children(child string) []formatter.IEntry {
	switch child {
	case proxiesChild:
		fallthrough
	default:
		length := len(e.ctx.Proxies())
		entries := make([]formatter.IEntry, length)
		for i := 0; i <= length; i++ {
			entries[length] = NewChildEntry(e, i, "proxy_", proxyFields)
		}
		return entries
	}
}

type ChildEntry struct {
	parent      *Entry
	index       int
	pre         string
	childReader IReaderIndex
}

func (c *ChildEntry) Read(pattern string) string {
	if strings.HasPrefix(pattern, c.pre) {
		name := strings.TrimPrefix(pattern, c.pre)
		v, _ := c.childReader.ReadByIndex(c.index, name, c.parent.ctx)
		return v
	}
	return c.parent.Read(pattern)
}

func (c *ChildEntry) Children(child string) []formatter.IEntry {
	return nil
}

func NewChildEntry(parent *Entry, index int, pre string, ReaderIndex IReaderIndex) *ChildEntry {
	return &ChildEntry{parent: parent, index: index, pre: pre}
}
