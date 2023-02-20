package metric_entry

import (
	"fmt"
	http_entry "github.com/eolinker/apinto/entries/http-entry"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
)

type metricEntry struct {
	iEntry  eosc.IEntry
	httpCtx http_context.IHttpContext
}

func NewMetricEntry(context eocontext.EoContext) (eosc.IMetricEntry, error) {
	switch v := context.(type) {
	case http_context.IHttpContext:
		return &metricEntry{
			iEntry:  http_entry.NewEntry(v),
			httpCtx: v,
		}, nil

	default:
		return nil, fmt.Errorf("NewMetricEntry fail. Unsupport type: %T", v)
	}

}

func (p *metricEntry) GetFloat(pattern string) (float64, bool) {
	f, exist := reqColRead[pattern]
	if !exist {
		log.Error("missing function belong to ", pattern)
		return 0, false
	}
	return f(p.httpCtx)
}

func (p *metricEntry) Read(pattern string) string {
	f, exist := reqLabelRead[pattern]
	if !exist {
		label := p.httpCtx.GetLabel(pattern)
		if label == "" {
			label = "-"
		}
		return label
	}
	return f(p.httpCtx)
}

func (p *metricEntry) Children(child string) []eosc.IMetricEntry {
	switch child {
	case http_entry.ProxiesChild:
		fallthrough
	default:
		proxies := p.httpCtx.Proxies()
		length := len(proxies)
		entries := make([]eosc.IMetricEntry, 0, length)
		for _, proxy := range proxies {
			entries = append(entries, newProxyMetricEntry(p.iEntry, proxy))
		}
		return entries
	}
}

type proxyMetricEntry struct {
	entry eosc.IEntry
	proxy http_context.IProxy
}

func (p *proxyMetricEntry) GetFloat(pattern string) (float64, bool) {
	f, exist := proxyColRead[pattern]
	if !exist {
		log.Error("missing function belong to ", pattern)
		return 0, false
	}
	return f(p.proxy)
}

func (p *proxyMetricEntry) Read(pattern string) string {
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

func (p proxyMetricEntry) Children(child string) []eosc.IMetricEntry {
	return nil
}

func newProxyMetricEntry(parent eosc.IEntry, proxy http_context.IProxy) eosc.IMetricEntry {
	return &proxyMetricEntry{
		entry: parent,
		proxy: proxy,
	}
}
