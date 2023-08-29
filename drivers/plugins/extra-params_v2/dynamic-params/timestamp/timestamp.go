package timestamp

import (
	"strconv"
	"time"

	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

const (
	defaultValue = "string"
)

type Timestamp struct {
	name  string
	value string
}

func NewTimestamp(name, value string) *Timestamp {
	return &Timestamp{name: name, value: value}
}

func (t *Timestamp) Name() string {
	return t.name
}

func (t *Timestamp) Generate(ctx http_service.IHttpContext, contentType string, args ...interface{}) (interface{}, error) {
	switch t.value {
	case "string":
		return strconv.FormatInt(time.Now().Unix(), 10), nil
	case "int":
		return time.Now().Unix(), nil
	}
	return strconv.FormatInt(time.Now().Unix(), 10), nil
}
