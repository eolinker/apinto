package fuse_strategy

import (
	"context"
	"fmt"
	"github.com/eolinker/apinto/utils/response"
	"strconv"
	"time"

	"github.com/eolinker/apinto/metrics"
	"github.com/eolinker/apinto/resources"
	"github.com/eolinker/apinto/strategy"
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
	return getFuseStatus(ctx, metrics, cache) == fuseStatusFusing
}

// 熔断次数的key
func getFuseCountKey(metrics string) string {
	return fmt.Sprintf("strategy-fuse:count:%s_%d", metrics, time.Now().Unix())
}

// 失败次数的key
func getErrorCountKey(metrics string) string {
	return fmt.Sprintf("strategy-fuse:error_count:%s_%d", metrics, time.Now().Unix())
}

func getSuccessCountKey(metrics string) string {
	return fmt.Sprintf("strategy-fuse:success_count:%s_%d", metrics, time.Now().Unix())
}
func getFuseStatusKey(metrics string) string {
	return fmt.Sprintf("strategy-fuse:status:%s", metrics)
}

func getFuseStatus(ctx context.Context, metrics string, cache resources.ICache) fuseStatus {

	key := getFuseStatusKey(metrics)
	expUnixStr, err := cache.Get(ctx, key).Result()
	if err != nil { //拿不到默认健康期
		return fuseStatusHealthy
	}

	expUnix, _ := strconv.ParseInt(expUnixStr, 16, 64)

	//过了熔断期是观察期
	if time.Now().UnixNano() > expUnix {
		return fuseStatusObserve
	}
	return fuseStatusFusing
}

type ruleHandler struct {
	metric                metrics.Metrics //熔断维度
	fuseConditionCount    int64
	fuseTime              fuseTimeConf
	recoverConditionCount int64
	response              response.IResponse
	codeStatusMap         map[int]codeStatus
}

type fuseTimeConf struct {
	time    time.Duration
	maxTime time.Duration
}

func NewFuseHandler(conf *Config) (*FuseHandler, error) {
	filter, err := strategy.ParseFilter(conf.Filters)
	if err != nil {
		return nil, err
	}

	codeStatusMap := make(map[int]codeStatus)
	for _, code := range conf.Rule.RecoverCondition.StatusCodes {
		codeStatusMap[code] = codeStatusSuccess
	}

	for _, code := range conf.Rule.FuseCondition.StatusCodes {
		codeStatusMap[code] = codeStatusError
	}
	rule := &ruleHandler{
		metric:             metrics.Parse([]string{conf.Rule.Metric}),
		fuseConditionCount: conf.Rule.FuseCondition.Count,

		fuseTime: fuseTimeConf{
			time:    time.Duration(conf.Rule.FuseTime.Time) * time.Second,
			maxTime: time.Duration(conf.Rule.FuseTime.MaxTime) * time.Second,
		},
		recoverConditionCount: conf.Rule.RecoverCondition.Count,

		response:      response.Parse(conf.Rule.Response),
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
