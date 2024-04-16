package limiting_strategy

import (
	"errors"
	http_entry "github.com/eolinker/apinto/entries/http-entry"
	"github.com/eolinker/apinto/utils/response"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
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
	entry := http_entry.NewEntry(httpContext)
	for _, h := range handlers {
		if h.Filter().Check(ctx) {
			key := h.Metrics().Key()
			if metricsAlready.Has(key) {
				continue
			}
			metricsAlready.Add(key)
			metricsValue := h.Metrics().Metrics(entry)

			if h.query.Second > 0 && scalars.QuerySecond.Get(metricsValue) >= h.query.Second {

				setLimitingStrategyContent(httpContext, h.Name(), h.Response())
				log.DebugF("refuse by limiting strategy %s of second query ", h.Name())

				return ErrorLimitingRefuse
			}
			if h.query.Minute > 0 && scalars.QueryMinute.Get(metricsValue) >= h.query.Minute {

				setLimitingStrategyContent(httpContext, h.Name(), h.Response())
				log.DebugF("refuse by limiting strategy %s of minute query ", h.Name())
				return ErrorLimitingRefuse
			}

			if h.query.Hour > 0 && scalars.QueryHour.Get(metricsValue) >= h.query.Hour {
				setLimitingStrategyContent(httpContext, h.Name(), h.Response())
				log.DebugF("refuse by limiting strategy %s of hour query ", h.Name())

				return ErrorLimitingRefuse
			}
			if h.traffic.Second > 0 && scalars.TrafficsSecond.Get(metricsValue) >= h.traffic.Second {

				setLimitingStrategyContent(httpContext, h.Name(), h.Response())
				log.DebugF("refuse by limiting strategy %s of second traffic ", h.Name())
				return ErrorLimitingRefuse
			}
			if h.traffic.Minute > 0 && scalars.TrafficsMinute.Get(metricsValue) >= h.traffic.Minute {
				setLimitingStrategyContent(httpContext, h.Name(), h.Response())
				log.DebugF("refuse by limiting strategy %s of minute traffic ", h.Name())
				return ErrorLimitingRefuse
			}

			if h.traffic.Hour > 0 && scalars.TrafficsHour.Get(metricsValue) >= h.traffic.Hour {
				setLimitingStrategyContent(httpContext, h.Name(), h.Response())
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
