package datetime

import (
	"time"

	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

const (
	defaultValue = "2006-01-02 15:04:05"
)

type Datetime struct {
	name  string
	value string
}

func NewDatetime(name string, value string) *Datetime {
	return &Datetime{name: name, value: value}
}

func (d *Datetime) Name() string {
	return d.name
}

func (d *Datetime) Generate(ctx http_service.IHttpContext, contentType string, args ...interface{}) (interface{}, error) {
	return time.Now().Format(d.value), nil
}
