package redis

import (
	"context"
	"errors"
	"time"

	"github.com/eolinker/apinto/resources"
)

var (
	ErrorNotInitRedis   = errors.New("redis not init")
	intError            = resources.NewIntResult(0, ErrorNotInitRedis)
	boolError           = resources.NewBoolResult(false, ErrorNotInitRedis)
	stringError         = resources.NewStringResult("", ErrorNotInitRedis)
	statusError         = resources.NewStatusResult(ErrorNotInitRedis)
	interfaceError      = resources.NewInterfaceResult(nil, ErrorNotInitRedis)
	arrayInterfaceError = resources.NewArrayInterfaceResult(nil, ErrorNotInitRedis)
)

type Empty struct {
}

func (e *Empty) BuildVector(name string, uni, step time.Duration) (resources.Vector, error) {
	return nil, ErrorNotInitRedis
}

func (e *Empty) Exec(ctx context.Context) error {
	return ErrorNotInitRedis
}

func (e *Empty) Set(ctx context.Context, key string, value []byte, expiration time.Duration) resources.StatusResult {

	return statusError
}

func (e *Empty) SetNX(ctx context.Context, key string, value []byte, expiration time.Duration) resources.BoolResult {
	return boolError
}

func (e *Empty) DecrBy(ctx context.Context, key string, decrement int64, expiration time.Duration) resources.IntResult {
	return intError
}

func (e *Empty) IncrBy(ctx context.Context, key string, decrement int64, expiration time.Duration) resources.IntResult {
	return intError
}

func (e *Empty) Get(ctx context.Context, key string) resources.StringResult {
	return stringError
}

func (e *Empty) GetDel(ctx context.Context, key string) resources.StringResult {
	return stringError
}

func (e *Empty) Del(ctx context.Context, keys ...string) resources.IntResult {
	return intError
}

func (e *Empty) HMSetN(ctx context.Context, key string, fields map[string]interface{}, expiration time.Duration) resources.BoolResult {
	return boolError
}

func (e *Empty) HMGet(ctx context.Context, key string, fields ...string) resources.ArrayInterfaceResult {
	return arrayInterfaceError
}

func (e *Empty) Run(ctx context.Context, script interface{}, keys []string, args ...interface{}) resources.InterfaceResult {
	return interfaceError
}

func (e *Empty) Tx() resources.TX {
	return e
}
