package uuid

import (
	"github.com/google/uuid"

	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

const (
	defaultValue = "string"
)

type Uuid struct {
	name  string
	value string
}

func NewUuid(name, value string) *Uuid {
	return &Uuid{name: name, value: value}
}

func (t *Uuid) Name() string {
	return t.name
}

func (t *Uuid) Generate(ctx http_service.IHttpContext, contentType string, args ...interface{}) (interface{}, error) {
	switch t.value {
	case "string":
		return uuid.New().String(), nil
	case "int":
		return uuid.New().ID(), nil
	}
	return uuid.New().String(), nil
}
