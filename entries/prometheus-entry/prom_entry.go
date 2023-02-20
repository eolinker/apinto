package prometheus_entry

import (
	"fmt"
	http_entry "github.com/eolinker/apinto/entries/http-entry"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
)

type metricEntry struct {
	entry     eosc.IEntry
	eoContext eocontext.EoContext
}

func NewMetricEntry(context interface{}) (eosc.IMetricEntry, error) {
	switch v := context.(type) {
	case http_context.IHttpContext:
		eoCtx := v.(eocontext.EoContext)
		return &metricEntry{
			entry:     http_entry.NewEntry(v),
			eoContext: eoCtx,
		}, nil

	default:
		return nil, fmt.Errorf("NewMetricEntry fail. Unsupport type: %T", v)
	}

}

func (p *metricEntry) GetFloat(pattern string) (float64, bool) {
	f, exist := reqColRead[pattern]
	if !exist {
		log.Error("missing function belong to ", pattern)
		return 0
	}
	return f(p.eoContext)
}

func (p *metricEntry) Read(pattern string) string {
	f, exist := reqLabelRead[pattern]
	if !exist {
		label := p.eoContext.GetLabel(pattern)
		if label == "" {
			label = "-"
		}
		return label
	}
	return f(p.eoContext)
}

func (p *metricEntry) Children(child string) []eosc.IMetricEntry {
	ctxProxies := p.entry.Children(child)

	proxyEntries := make([]eosc.IMetricEntry, 0, len(ctxProxies))

	for _, proxy := range ctxProxies {
		proxyEntries = append(proxyEntries, newProxyPromEntry(p.entry, proxy))
	}

	return proxyEntries
}

type proxyPromEntry struct {
	entry eosc.IEntry
	proxy http_context.IProxy
}

func (p *proxyPromEntry) GetFloat(pattern string) (float64, bool) {
	f, exist := proxyColRead[pattern]
	if !exist {
		log.Error("missing function belong to ", pattern)
		return 0
	}
	return f(p.proxy)
}

func (p *proxyPromEntry) Read(pattern string) string {
	f, exist := proxyLabelRead[pattern]
	if !exist {
		label := p.parent.context.GetLabel(pattern)
		if label == "" {
			label = "-"
		}
		return label
	}
	return f(p.proxy)
}

func (p proxyPromEntry) Children(child string) []eosc.IMetricEntry {
	return nil
}

func newProxyPromEntry(parent eosc.IEntry, proxy http_context.IProxy) *proxyPromEntry {
	return &proxyPromEntry{
		entry: parent,
		proxy: proxy,
	}
}
