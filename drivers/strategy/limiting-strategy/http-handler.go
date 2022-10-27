package limiting_strategy

import (
	"errors"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
	"net/http"
	"strconv"
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

func (hd *actuatorHttp) Check(ctx eocontext.EoContext, handlers []*LimitingHandler, scalars *Scalars) error {
	httpContext, err := http_service.Assert(ctx)
	if err != nil {
		return err
	}
	contentLength, _ := strconv.ParseInt(httpContext.Request().Header().GetHeader("content-length"), 10, 64)

	metricsAlready := newSet(len(handlers))
	for _, h := range handlers {
		if h.Filter().Check(ctx) {
			key := h.Metrics().Key()
			if metricsAlready.Has(key) {
				continue
			}
			metricsAlready.Add(key)
			metricsValue := h.Metrics().Metrics(ctx)

			if scalars.QuerySecond.Get(metricsValue) > h.query.Second {

				setLimitingStrategyContent(httpContext, h.Name())
				log.DebugF("refuse by limiting strategy %s of second query ", h.Name())

				return ErrorLimitingRefuse
			}
			if scalars.QueryMinute.Get(metricsValue) > h.query.Minute {

				setLimitingStrategyContent(httpContext, h.Name())
				log.DebugF("refuse by limiting strategy %s of minute query ", h.Name())
				return ErrorLimitingRefuse
			}

			if scalars.QueryMinute.Get(metricsValue) > h.query.Hour {
				setLimitingStrategyContent(httpContext, h.Name())
				log.DebugF("refuse by limiting strategy %s of hour query ", h.Name())

				return ErrorLimitingRefuse
			}
			if scalars.TrafficsSecond.Get(metricsValue) > h.traffic.Second {

				setLimitingStrategyContent(httpContext, h.Name())
				log.DebugF("refuse by limiting strategy %s of second traffic ", h.Name())
				return ErrorLimitingRefuse
			}
			if scalars.TrafficsMinute.Get(metricsValue) > h.traffic.Minute {
				setLimitingStrategyContent(httpContext, h.Name())
				log.DebugF("refuse by limiting strategy %s of minute traffic ", h.Name())
				return ErrorLimitingRefuse
			}

			if scalars.TrafficsHour.Get(metricsValue) > h.traffic.Hour {
				setLimitingStrategyContent(httpContext, h.Name())
				log.DebugF("refuse by limiting strategy %s of hour traffic ", h.Name())

				return ErrorLimitingRefuse
			}
			scalars.QuerySecond.Add(metricsValue, 1)
			scalars.QueryMinute.Add(metricsValue, 1)
			scalars.QueryHour.Add(metricsValue, 1)
			scalars.TrafficsSecond.Add(metricsValue, contentLength)
			scalars.TrafficsMinute.Add(metricsValue, contentLength)
			scalars.TrafficsHour.Add(metricsValue, contentLength)

		}
	}
	return nil
}
func setLimitingStrategyContent(httpContext http_service.IHttpContext, name string) {
	httpContext.Response().SetStatus(http.StatusForbidden, http.StatusText(http.StatusForbidden))
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
