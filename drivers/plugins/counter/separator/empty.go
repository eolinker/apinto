package separator

import http_service "github.com/eolinker/eosc/eocontext/http-context"

type EmptyCounter struct {
}

func NewEmptyCounter() *EmptyCounter {
	return &EmptyCounter{}
}

func (e *EmptyCounter) Count(ctx http_service.IHttpContext) (int64, error) {
	return 1, nil
}

func (e *EmptyCounter) Max() int64 {
	return 2
}

func (e *EmptyCounter) Name() string {
	return "__empty"
}
