package nsq

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/eolinker/eosc/log"

	"github.com/ohler55/ojg/oj"

	"github.com/eolinker/eosc"
	"github.com/ohler55/ojg/jp"
)

var (
	pushModeMulti = "multi"
)

type CounterHandler struct {
	counters  eosc.Untyped[string, ICounter]
	pushMode  string
	paramExpr []*paramExpr
	countExpr jp.Expr
}

func newCounterHandler(params []*Param, countParamKey string, pushMode string) (*CounterHandler, error) {
	pes := make([]*paramExpr, 0, len(params))
	for _, p := range params {
		if p.Key == "" {
			continue
		}
		key := p.Key
		if !strings.HasPrefix(key, "$.") {
			key = "$." + key
		}
		expr, err := jp.ParseString(key)
		if err != nil {
			return nil, fmt.Errorf("parse param key %s error: %w", p.Key, err)
		}
		pes = append(pes, &paramExpr{
			expr:       expr,
			value:      p.Value,
			isVariable: strings.HasPrefix(p.Value, "$"),
			typ:        p.Type,
		})
	}
	if countParamKey == "" {
		countParamKey = "$." + countParamKey
	}
	return &CounterHandler{
		counters:  eosc.BuildUntyped[string, ICounter](),
		pushMode:  pushMode,
		paramExpr: pes,
		countExpr: jp.MustParseString(countParamKey),
	}, nil
}

func (c *CounterHandler) Generate() ([][]byte, error) {
	counters := c.counters.All()
	switch c.pushMode {
	case pushModeMulti:
		result := make([]interface{}, 0, len(counters))
		for _, ct := range counters {
			count := ct.Count()
			if count < 1 {
				continue
			}
			body, err := ct.Generate(count)
			if err != nil {
				log.Error(err)
				continue
			}
			result = append(result, body)
		}
		if len(result) < 1 {
			return nil, nil
		}
		data, _ := oj.Marshal(result)
		return [][]byte{data}, nil
	default:
		result := make([][]byte, 0, len(counters))
		for _, ct := range counters {
			count := ct.Count()
			if count < 1 {
				continue
			}
			body, err := ct.Generate(count)
			if err != nil {
				log.Error(err)
				continue
			}
			data, _ := oj.Marshal(body)
			result = append(result, data)
		}
		return result, nil
	}
}

func (c *CounterHandler) GetCounter(key string, variables map[string]string) ICounter {
	counter, ok := c.counters.Get(key)
	if ok {
		return counter
	}
	// 不存在则创建
	body, _ := oj.Parse([]byte("{}"))
	for _, expr := range c.paramExpr {
		err := expr.expr.Set(body, expr.GetValue(variables))
		if err != nil {
			log.Error(err)
			continue
		}
	}
	counter = NewCounter(key, body, c.countExpr)
	c.counters.Set(key, counter)
	return counter
}

// ICounter 计数器
type ICounter interface {
	Key() string
	Add(count int64)
	Count() int64
	// Generate 返回计数器生成的信息，同时清理计数器
	Generate(count int64) (interface{}, error)
}

type Counter struct {
	key           string
	count         int64
	body          interface{}
	timestampExpr jp.Expr
	datetimeExpr  jp.Expr
	countExpr     jp.Expr
}

func NewCounter(key string, body interface{}, countExpr jp.Expr) *Counter {
	return &Counter{
		key:           key,
		body:          body,
		countExpr:     countExpr,
		timestampExpr: jp.MustParseString("$.timestamp"),
		datetimeExpr:  jp.MustParseString("$.datetime"),
	}
}

func (c *Counter) Key() string {
	return c.key
}

func (c *Counter) Add(count int64) {
	atomic.AddInt64(&c.count, count)
}

func (c *Counter) Count() int64 {
	// 读出旧值，同时将值赋0
	return atomic.SwapInt64(&c.count, 0)
}

func (c *Counter) Generate(count int64) (interface{}, error) {
	err := c.countExpr.Set(c.body, count)
	if err != nil {
		// 当body设置失败，此时count值已经丢失，需要将count值重新加回去，返回报错
		atomic.AddInt64(&c.count, count)
		orgBody, _ := oj.Marshal(c.body)
		return nil, fmt.Errorf("set count to body failed,err: %s,body: %s", err.Error(), string(orgBody))
	}
	now := time.Now()
	c.timestampExpr.Set(c.body, now.UnixMicro())
	c.datetimeExpr.Set(c.body, now.Format("2006-01-02 15:04:05"))

	return c.body, nil
}
