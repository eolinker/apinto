package router_http

import "net/url"

type TargetConfig struct {
	location string
	host     string
	header   map[string]string
	query    url.Values
}

func (t *TargetConfig) Location() string {
	return t.location
}

func (t *TargetConfig) Host() string {
	return t.host
}

func (t *TargetConfig) Header() map[string]string {
	return t.header
}

func (t *TargetConfig) Query() url.Values {
	return t.query
}
