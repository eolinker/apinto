package fuse_strategy

import (
	"fmt"
	"github.com/coocood/freecache"
	"github.com/eolinker/apinto/metrics"
	"github.com/eolinker/apinto/resources"
	"github.com/eolinker/apinto/strategy"
	"github.com/eolinker/eosc/eocontext"
	"github.com/go-redis/redis/v8"
	"time"
)

type fuseStatus string

const (
	fuseStatusHealthy fuseStatus = "healthy" //健康期间
	fuseStatusFusing  fuseStatus = "fusing"  //熔断期间
	fuseStatusObserve fuseStatus = "observe" //观察期
)

type FuseHandler struct {
	name     string
	filter   strategy.IFilter
	priority int
	stop     bool
	rule     *ruleHandler
	status   fuseStatus //状态
}

func (f *FuseHandler) IsFuse(eoCtx eocontext.EoContext, cache resources.ICache) bool {
	return f.getFuseStatus(eoCtx, cache) == fuseStatusFusing
}

func (f *FuseHandler) getFuseCountKey(eoCtx eocontext.EoContext) string {
	return fmt.Sprintf("fuse_%s_%d", f.rule.metric.Metrics(eoCtx), time.Now().Second())
}

func (f *FuseHandler) getRecoverCountKey(eoCtx eocontext.EoContext) string {
	return fmt.Sprintf("fuse_recover_%s_%d", f.rule.metric.Metrics(eoCtx), time.Now().Second())
}

func (f *FuseHandler) getFuseTimeKey(eoCtx eocontext.EoContext) string {
	return fmt.Sprintf("fuse_time_%s", f.rule.metric.Metrics(eoCtx))
}

func (f *FuseHandler) getFuseStatus(eoCtx eocontext.EoContext, cache resources.ICache) fuseStatus {

	ctx := eoCtx.Context()

	key := fmt.Sprintf("fuse_status_%s", f.rule.metric.Metrics(eoCtx))
	status, err := cache.Get(ctx, key).Result()
	if err != nil { //拿不到默认健康期
		return fuseStatusHealthy
	}

	_, err = cache.Get(ctx, f.getFuseTimeKey(eoCtx)).Result()
	if err != nil && (err == freecache.ErrNotFound || err == redis.Nil) && fuseStatus(status) == fuseStatusFusing { //记录的状态如果是熔断期，此时熔断时间又过期了，则返回观察期
		return fuseStatusObserve
	}

	return fuseStatus(status)
}

func (f *FuseHandler) setFuseStatus(eoCtx eocontext.EoContext, cache resources.ICache, status fuseStatus, expiration time.Duration) {
	key := fmt.Sprintf("fuse_status_%s", f.rule.metric.Metrics(eoCtx))
	cache.Set(eoCtx.Context(), key, []byte(status), expiration)
}

type ruleHandler struct {
	metric           metrics.Metrics //熔断维度
	fuseCondition    statusConditionConf
	fuseTime         fuseTimeConf
	recoverCondition statusConditionConf
	response         strategyResponseConf
}

type statusConditionConf struct {
	statusCodes []int
	count       int64
}

type fuseTimeConf struct {
	time    int64
	maxTime int64
}

type strategyResponseConf struct {
	statusCode  int
	contentType string
	charset     string
	headers     []header
	body        string
}
type header struct {
	key   string
	value string
}

func NewFuseHandler(conf *Config) (*FuseHandler, error) {
	filter, err := strategy.ParseFilter(conf.Filters)
	if err != nil {
		return nil, err
	}

	headers := make([]header, 0)
	for _, v := range conf.Rule.Response.Header {
		headers = append(headers, header{
			key:   v.key,
			value: v.value,
		})
	}

	rule := &ruleHandler{
		metric: metrics.Parse([]string{conf.Rule.Metric}),
		fuseCondition: statusConditionConf{
			statusCodes: conf.Rule.FuseCondition.StatusCodes,
			count:       conf.Rule.FuseCondition.Count,
		},
		fuseTime: fuseTimeConf{
			time:    conf.Rule.FuseTime.Time,
			maxTime: conf.Rule.FuseTime.MaxTime,
		},
		recoverCondition: statusConditionConf{
			statusCodes: conf.Rule.RecoverCondition.StatusCodes,
			count:       conf.Rule.RecoverCondition.Count,
		},
		response: strategyResponseConf{
			statusCode:  conf.Rule.Response.StatusCode,
			contentType: conf.Rule.Response.ContentType,
			charset:     conf.Rule.Response.Charset,
			headers:     headers,
			body:        conf.Rule.Response.Body,
		},
	}
	return &FuseHandler{
		name:     conf.Name,
		filter:   filter,
		priority: conf.Priority,
		stop:     conf.Stop,
		rule:     rule,
	}, nil
}
