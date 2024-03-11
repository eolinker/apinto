package metric_entry

import (
	"fmt"
	"strings"

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
		return nil, fmt.Errorf("NewMetricEntry fail. Unsupported type: %T", v)
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
	//会先从rule里面读，若rule没有相应的pattern，会从ctx里面读label
	value := eosc.ReadStringFromEntry(p.iEntry, pattern)
	if value == "" {
		value = "-"
	}

	return value
}

func (p *metricEntry) Children(child string) []eosc.IMetricEntry {
	switch child {
	case http_entry.ProxiesChild:
		fallthrough
	default:
		p.iEntry.Children(child)
		proxies := p.httpCtx.Proxies()
		length := len(proxies)
		entries := make([]eosc.IMetricEntry, 0, length)
		for _, proxy := range proxies {
			entries = append(entries, newProxyMetricEntry(p, proxy, "proxy_"))
		}
		return entries
	}
}

type proxyMetricEntry struct {
	parent eosc.IMetricEntry
	proxy  http_context.IProxy

	prefix       string
	proxyReaders http_entry.ProxyReaders
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
	if strings.HasPrefix(pattern, p.prefix) {
		name := strings.TrimPrefix(pattern, p.prefix)
		f, exist := p.proxyReaders[name]
		if exist {
			value, has := http_entry.ReadProxyFromProxyReader(f, p.proxy, name)
			if !has {
				value = "-"
			}
			return value
		}
	}

	return p.parent.Read(pattern)
}

func (p proxyMetricEntry) Children(child string) []eosc.IMetricEntry {
	return nil
}

func newProxyMetricEntry(parent eosc.IMetricEntry, proxy http_context.IProxy, prefix string) eosc.IMetricEntry {
	return &proxyMetricEntry{
		parent:       parent,
		proxy:        proxy,
		prefix:       prefix,
		proxyReaders: http_entry.GetProxyReaders(),
	}
}
