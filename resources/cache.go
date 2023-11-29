package resources

import (
	"context"
	"time"
	"unsafe"
)

const CacheSkill = "github.com/eolinker/apinto/resources.resources.ICache"

type ICache interface {
	Set(ctx context.Context, key string, value []byte, expiration time.Duration) StatusResult
	SetNX(ctx context.Context, key string, value []byte, expiration time.Duration) BoolResult
	DecrBy(ctx context.Context, key string, decrement int64, expiration time.Duration) IntResult
	IncrBy(ctx context.Context, key string, decrement int64, expiration time.Duration) IntResult
	Get(ctx context.Context, key string) StringResult
	GetDel(ctx context.Context, key string) StringResult
	Del(ctx context.Context, keys ...string) IntResult
	Run(ctx context.Context, script interface{}, keys []string, args ...interface{}) InterfaceResult
	Tx() TX
}

type TX interface {
	ICache
	Exec(ctx context.Context) error
}

type InterfaceResult interface {
	Result() (interface{}, error)
}

type BoolResult interface {
	Result() (bool, error)
}

type IntResult interface {
	Result() (int64, error)
}

type StringResult interface {
	Result() (string, error)
	Bytes() ([]byte, error)
}

type StatusResult interface {
	Result() error
}

type statusResult struct {
	err error
}

func NewStatusResult(err error) *statusResult {
	return &statusResult{err: err}
}

func (s *statusResult) Result() error {
	return s.err
}

type stringResult struct {
	err error
	val string
}

func NewStringResult(val string, err error) *stringResult {
	return &stringResult{err: err, val: val}
}

func NewStringResultBytes(value []byte, err error) *stringResult {
	return &stringResult{val: *(*string)(unsafe.Pointer(&value)), err: err}
}

func (s *stringResult) Result() (string, error) {
	return s.val, s.err
}

func (s *stringResult) Bytes() ([]byte, error) {
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			Cap int
		}{s.val, len(s.val)},
	)), s.err
}

type boolResult struct {
	val bool
	err error
}

func NewBoolResult(val bool, err error) *boolResult {
	return &boolResult{val: val, err: err}
}

func (b *boolResult) Result() (bool, error) {
	return b.val, b.err
}

type intResult struct {
	val int64
	err error
}

func NewIntResult(val int64, err error) *intResult {
	return &intResult{val: val, err: err}
}

func (b *intResult) Result() (int64, error) {
	return b.val, b.err
}

type interfaceResult struct {
	val interface{}
	err error
}

func NewInterfaceResult(val interface{}, err error) *interfaceResult {
	return &interfaceResult{val: val, err: err}
}

func (b *interfaceResult) Result() (interface{}, error) {
	return b.val, b.err
}
