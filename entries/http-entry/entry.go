package http_entry

import (
	"strings"

	"github.com/eolinker/eosc"

	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

var (
	_            eosc.IEntry = (*Entry)(nil)
	_            eosc.IEntry = (*ChildEntry)(nil)
	ProxiesChild             = "proxies"
)

type Entry struct {
	ctx http_service.IHttpContext
}

func (e *Entry) ReadLabel(pattern string) string {
	return eosc.String(e.Read(pattern))
}

func NewEntry(ctx http_service.IHttpContext) eosc.IEntry {
	return &Entry{ctx: ctx}
}

func (e *Entry) Read(pattern string) interface{} {
	v, ok := rule.Read(pattern, e.ctx)
	if !ok {
		return ""
	}

	return v
}

func (e *Entry) Children(child string) []eosc.IEntry {
	switch child {
	case ProxiesChild:
		fallthrough
	default:
		length := len(e.ctx.Proxies())
		entries := make([]eosc.IEntry, length)
		for i := 0; i < length; i++ {
			entries[i] = NewChildEntry(e, i, "proxy_", proxyFields)
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

func (c *ChildEntry) ReadLabel(pattern string) string {
	return eosc.String(c.Read(pattern))
}

func (c *ChildEntry) Read(pattern string) interface{} {
	if strings.HasPrefix(pattern, c.pre) {
		name := strings.TrimPrefix(pattern, c.pre)
		v, _ := c.childReader.ReadByIndex(c.index, name, c.parent.ctx)
		return v
	}
	return c.parent.Read(pattern)
}

func (c *ChildEntry) Children(child string) []eosc.IEntry {
	return nil
}

func NewChildEntry(parent *Entry, index int, pre string, ReaderIndex IReaderIndex) *ChildEntry {
	return &ChildEntry{parent: parent, index: index, pre: pre, childReader: ReaderIndex}
}
