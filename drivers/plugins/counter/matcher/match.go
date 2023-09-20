package matcher

import (
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

type IMatcher interface {
	Match(ctx http_service.IHttpContext) bool
}

type MatchParam struct {
	Key   string   `json:"key"`
	Kind  string   `json:"kind" enum:"int,string,bool" default:"string"` // int|string|bool
	Value []string `json:"value"`
}
