package prometheus_entry

import (
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

type IPromEntry interface {
	GetValue(key string) float64
	GetLabel(key string) string
	Proxy() []IPromEntry
}

type promEntry struct {
	context http_service.IHttpContext
}

func NewPromEntry(context http_service.IHttpContext) IPromEntry {
	return &promEntry{
		context: context,
	}
}

func (p *promEntry) GetValue(key string) float64 {
	//TODO implement me
	panic("implement me")
}

func (p *promEntry) GetLabel(key string) string {
	//TODO implement me
	panic("implement me")
}

func (p *promEntry) Proxy() []IPromEntry {
	//TODO implement me
	panic("implement me")
}

type proxyPromEntry struct {
	parent *promEntry
	//childReader IReaderIndex
}

func (p *proxyPromEntry) GetValue(key string) float64 {
	//TODO implement me
	panic("implement me")
}

func (p *proxyPromEntry) GetLabel(key string) string {
	//TODO implement me
	panic("implement me")
}

func (p proxyPromEntry) Proxy() []IPromEntry {
	//TODO implement me
	panic("implement me")
}

func NewProxyPromEntry(parent *promEntry) *proxyPromEntry {
	return &proxyPromEntry{parent: parent}
}
