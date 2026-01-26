package limiting_strategy

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/eolinker/apinto/resources"

	http_entry "github.com/eolinker/apinto/entries/http-entry"
	"github.com/eolinker/apinto/utils/response"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
)

func init() {
	RegisterActuator(NewActuator())
}

var (
	ErrorLimitingRefuse = errors.New("refuse by limiting strategy")
)

type actuatorHttp struct {
}

func NewActuator() *actuatorHttp {
	return &actuatorHttp{}
}

func (hd *actuatorHttp) Assert(ctx eocontext.EoContext) bool {
	_, err := http_service.Assert(ctx)
	if err != nil {
		return false
	}
	return true
}

func (hd *actuatorHttp) compareAndAddCount(ctx http_service.IHttpContext, vector resources.Vector, metricsValue, period, handlerName string, threshold int64, response response.IResponse) error {
	value := vector.Get(metricsValue)
	if value >= threshold {
		setLimitingStrategyContent(ctx, handlerName, response)
		log.DebugF("refuse by limiting strategy %s of %s query.", handlerName, period)
		ctx.WithValue("is_block", true)
		ctx.SetLabel("block_name", handlerName)
		return ErrorLimitingRefuse
	}
	ctx.Response().SetHeader(fmt.Sprintf("x-rate-limit-%s", period), fmt.Sprintf("%d/%d", value+1, threshold))
	vector.Add(metricsValue, 1)
	return nil
}

func (hd *actuatorHttp) compareAndAddLength(ctx http_service.IHttpContext, vector resources.Vector, metricsValue, period, handlerName string, threshold, contentLength int64, response response.IResponse) error {
	if vector.Get(metricsValue) >= threshold {
		setLimitingStrategyContent(ctx, handlerName, response)
		log.DebugF("refuse by limiting strategy %s of %s traffic.", handlerName, period)
		ctx.WithValue("is_block", true)
		ctx.SetLabel("block_name", handlerName)
		return ErrorLimitingRefuse
	}
	vector.Add(metricsValue, contentLength)
	return nil
}

func (hd *actuatorHttp) Check(ctx eocontext.EoContext, handlers []*LimitingHandler, scalars *Scalars) error {
	httpContext, err := http_service.Assert(ctx)
	if err != nil {
		return err
	}
	contentLength, _ := strconv.ParseInt(httpContext.Request().Header().GetHeader("content-length"), 10, 64)

	metricsAlready := newSet(len(handlers))
	entry := http_entry.NewEntry(httpContext)
	for _, h := range handlers {
		if h.Filter().Check(ctx) {
			key := h.Metrics().Key()
			if metricsAlready.Has(key) {
				continue
			}
			metricsAlready.Add(key)
			metricsValue := h.Metrics().Metrics(entry)

			if h.query.Second > 0 {
				err = hd.compareAndAddCount(httpContext, scalars.QuerySecond, metricsValue, "second", h.name, h.query.Second, h.Response())
				if err != nil {
					return err
				}
			}
			if h.query.Minute > 0 {
				err = hd.compareAndAddCount(httpContext, scalars.QueryMinute, metricsValue, "minute", h.name, h.query.Minute, h.Response())
				if err != nil {
					return err
				}
			}

			if h.query.Hour > 0 {
				err = hd.compareAndAddCount(httpContext, scalars.QueryHour, metricsValue, "hour", h.name, h.query.Hour, h.Response())
				if err != nil {
					return err
				}
			}

			if h.traffic.Second > 0 {
				err = hd.compareAndAddLength(httpContext, scalars.TrafficsSecond, metricsValue, "second", h.name, h.traffic.Second, contentLength, h.Response())
				if err != nil {
					return err
				}
			}
			if h.traffic.Minute > 0 {
				err = hd.compareAndAddLength(httpContext, scalars.TrafficsMinute, metricsValue, "minute", h.name, h.traffic.Minute, contentLength, h.Response())
				if err != nil {
					return err
				}
			}

			if h.traffic.Hour > 0 {
				err = hd.compareAndAddLength(httpContext, scalars.TrafficsHour, metricsValue, "hour", h.name, h.traffic.Hour, contentLength, h.Response())
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
func setLimitingStrategyContent(httpContext http_service.IHttpContext, name string, res response.IResponse) {
	res.Response(httpContext)
	httpContext.Response().SetHeader("strategy", name)
}

type Set map[string]struct{}

func newSet(l int) Set {
	return make(Set, l)
}
func (s Set) Has(key string) bool {
	_, has := s[key]
	return has
}
func (s Set) Add(key string) {
	s[key] = struct{}{}
}
