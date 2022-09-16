package http

import (
	"errors"
	limiting_stragety "github.com/eolinker/apinto/drivers/strategy/limiting-stragety"
	"github.com/eolinker/apinto/drivers/strategy/limiting-stragety/scalar"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
	"net/http"
	"strconv"
)

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

func (hd *actuatorHttp) Check(ctx eocontext.EoContext, handlers []*limiting_stragety.LimitingHandler, queryScalars scalar.Manager, trafficScalars scalar.Manager) error {
	httpContext, err := http_service.Assert(ctx)
	if err != nil {
		return err
	}
	length, _ := strconv.ParseUint(httpContext.Request().Header().GetHeader("content-length"), 10, 64)

	metricsAlready := newSet(len(handlers))
	for _, h := range handlers {
		if h.Filter().Check(ctx) {
			key := h.Metrics().Key()
			if metricsAlready.Has(key) {
				continue
			}
			metricsAlready.Add(key)
			metricsValue := h.Metrics().Metrics(ctx)

			queryScalar := queryScalars.Get(metricsValue)
			trafficScalar := trafficScalars.Get(metricsValue)
			if !queryScalar.Second().CompareAndAdd(h.Query().Second, 1) {
				setLimitingStrategyContent(httpContext, h.Name())
				log.DebugF("refuse by limiting strategy %s of second query ", h.Name())

				return ErrorLimitingRefuse
			}
			if !queryScalar.Minute().CompareAndAdd(h.Query().Minute, 1) {
				setLimitingStrategyContent(httpContext, h.Name())
				log.DebugF("refuse by limiting strategy %s of minute query ", h.Name())

				return ErrorLimitingRefuse
			}
			if !queryScalar.Hour().CompareAndAdd(h.Query().Hour, 1) {
				setLimitingStrategyContent(httpContext, h.Name())
				log.DebugF("refuse by limiting strategy %s of hour query ", h.Name())

				return ErrorLimitingRefuse
			}

			if !trafficScalar.Second().CompareAndAdd(h.Traffic().Second, length) {
				setLimitingStrategyContent(httpContext, h.Name())
				log.DebugF("refuse by limiting strategy %s of second traffic ", h.Name())
				return ErrorLimitingRefuse
			}

			if !trafficScalar.Minute().CompareAndAdd(h.Traffic().Minute, length) {
				setLimitingStrategyContent(httpContext, h.Name())
				log.DebugF("refuse by limiting strategy %s of minute traffic ", h.Name())
				return ErrorLimitingRefuse
			}

			if !trafficScalar.Hour().CompareAndAdd(h.Traffic().Hour, length) {
				setLimitingStrategyContent(httpContext, h.Name())
				log.DebugF("refuse by limiting strategy %s of hour traffic ", h.Name())

				return ErrorLimitingRefuse
			}
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
