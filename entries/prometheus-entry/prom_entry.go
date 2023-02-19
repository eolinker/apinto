package prometheus_entry

import (
	http_context "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
)

type IPromEntry interface {
	GetValue(key string) float64
	GetLabel(key string) string
	Proxies() []IPromEntry
}

type promEntry struct {
	context http_context.IHttpContext
}

func NewPromEntry(context http_context.IHttpContext) IPromEntry {
	return &promEntry{
		context: context,
	}
}

func (p *promEntry) GetValue(key string) float64 {
	f, exist := reqColRead[key]
	if !exist {
		log.Error("missing function belong to ", key)
		return 0
	}
	return f(p.context)
}

func (p *promEntry) GetLabel(key string) string {
	f, exist := reqLabelRead[key]
	if !exist {
		label := p.context.GetLabel(key)
		if label == "" {
			label = "-"
		}
		return label
	}
	return f(p.context)
}

func (p *promEntry) Proxies() []IPromEntry {
	ctxProxies := p.context.Proxies()

	proxyEntries := make([]IPromEntry, 0, len(ctxProxies))

	for _, proxy := range ctxProxies {
		proxyEntries = append(proxyEntries, newProxyPromEntry(p, proxy))
	}

	return proxyEntries
}

type proxyPromEntry struct {
	parent *promEntry
	proxy  http_context.IProxy
}

func (p *proxyPromEntry) GetValue(key string) float64 {
	f, exist := proxyColRead[key]
	if !exist {
		log.Error("missing function belong to ", key)
		return 0
	}
	return f(p.proxy)
}

func (p *proxyPromEntry) GetLabel(key string) string {
	f, exist := proxyLabelRead[key]
	if !exist {
		label := p.parent.context.GetLabel(key)
		if label == "" {
			label = "-"
		}
		return label
	}
	return f(p.proxy)
}

func (p proxyPromEntry) Proxies() []IPromEntry {
	return nil
}

func newProxyPromEntry(parent *promEntry, proxy http_context.IProxy) *proxyPromEntry {
	return &proxyPromEntry{
		parent: parent,
		proxy:  proxy,
	}
}
