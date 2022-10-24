package fuse_strategy

import (
	"fmt"
	"github.com/eolinker/apinto/metrics"
	"github.com/eolinker/apinto/resources"
	"github.com/eolinker/apinto/strategy"
	"golang.org/x/net/context"
	"strconv"
	"time"
)

type fuseStatus string

const (
	fuseStatusHealthy fuseStatus = "healthy" //健康期间
	fuseStatusFusing  fuseStatus = "fusing"  //熔断期间
	fuseStatusObserve fuseStatus = "observe" //观察期
)

type codeStatus int

const (
	codeStatusSuccess codeStatus = 1
	codeStatusError   codeStatus = 2
)

type FuseHandler struct {
	name     string
	filter   strategy.IFilter
	priority int
	stop     bool
	rule     *ruleHandler
}

func (f *FuseHandler) IsFuse(ctx context.Context, metrics string, cache resources.ICache) bool {
	return f.getFuseStatus(ctx, metrics, cache) == fuseStatusFusing
}

//熔断次数的key
func (f *FuseHandler) getFuseCountKey(metrics string) string {
	return fmt.Sprintf("fuse_count_%s_%d", metrics, time.Now().Second())
}

//失败次数的key
func (f *FuseHandler) getErrorCountKey(metrics string) string {
	return fmt.Sprintf("fuse_error_count_%s_%d", metrics, time.Now().Second())
}

func (f *FuseHandler) getSuccessCountKey(metrics string) string {
	return fmt.Sprintf("fuse_success_count_%s_%d", metrics, time.Now().Second())
}

func (f *FuseHandler) getFuseTimeKey(metrics string) string {
	return fmt.Sprintf("fuse_time_%s", metrics)
}

func (f *FuseHandler) getFuseStatus(ctx context.Context, metrics string, cache resources.ICache) fuseStatus {

	key := f.getFuseStatusKey(metrics)
	expUnixStr, err := cache.Get(ctx, key).Result()
	if err != nil { //拿不到默认健康期
		return fuseStatusHealthy
	}

	expUnix, _ := strconv.Atoi(expUnixStr)

	//过了熔断期是观察期
	if time.Now().Unix() > int64(expUnix) {
		return fuseStatusObserve
	}
	return fuseStatusFusing
}

func (f *FuseHandler) getFuseStatusKey(metrics string) string {
	return fmt.Sprintf("fuse_status_%s", metrics)
}

func (f *FuseHandler) setFuseStatus(ctx context.Context, metrics string, cache resources.ICache, exp string, expiration time.Duration) {
	key := f.getFuseStatusKey(metrics)
	cache.Set(ctx, key, []byte(exp), expiration)
}

func (f *FuseHandler) delFuseStatus(ctx context.Context, metrics string, cache resources.ICache) {
	key := f.getFuseStatusKey(metrics)
	cache.Del(ctx, key)
}

type ruleHandler struct {
	metric           metrics.Metrics //熔断维度
	fuseCondition    statusConditionConf
	fuseTime         fuseTimeConf
	recoverCondition statusConditionConf
	response         strategyResponseConf
	codeStatusMap    map[int]codeStatus
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

	codeStatusMap := make(map[int]codeStatus)
	for _, code := range conf.Rule.RecoverCondition.StatusCodes {
		codeStatusMap[code] = codeStatusSuccess
	}

	for _, code := range conf.Rule.FuseCondition.StatusCodes {
		codeStatusMap[code] = codeStatusError
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
		codeStatusMap: codeStatusMap,
	}
	return &FuseHandler{
		name:     conf.Name,
		filter:   filter,
		priority: conf.Priority,
		stop:     conf.Stop,
		rule:     rule,
	}, nil
}
