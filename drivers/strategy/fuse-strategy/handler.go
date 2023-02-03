package fuse_strategy

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/eolinker/apinto/metrics"
	"github.com/eolinker/apinto/resources"
	"github.com/eolinker/apinto/strategy"
	"golang.org/x/net/context"
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
	return fmt.Sprintf("fuse_count_%s_%d", metrics, time.Now().Unix())
}

// 失败次数的key
func getErrorCountKey(metrics string) string {
	return fmt.Sprintf("fuse_error_count_%s_%d", metrics, time.Now().Unix())
}

func getSuccessCountKey(metrics string) string {
	return fmt.Sprintf("fuse_success_count_%s_%d", metrics, time.Now().Unix())
}
func getFuseStatusKey(metrics string) string {
	return fmt.Sprintf("fuse_status_%s", metrics)
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
	response              strategyResponseConf
	codeStatusMap         map[int]codeStatus
}

type statusConditionConf struct {
	statusCodes []int
	count       int64
}

type fuseTimeConf struct {
	time    time.Duration
	maxTime time.Duration
}

type strategyResponseConf struct {
	statusCode  int
	contentType string
	charset     string
	headers     []header
	body        string
}

func (s *strategyResponseConf) Body(labels map[string]string) string {
	tempBody := s.body

	tempBody = strings.ReplaceAll(tempBody, "$api", fmt.Sprintf("%s(%s)", labels["api"], labels["api_id"]))
	tempBody = strings.ReplaceAll(tempBody, "$api_id", labels["api_id"])
	tempBody = strings.ReplaceAll(tempBody, "$api_name", labels["api"])

	tempBody = strings.ReplaceAll(tempBody, "$application", fmt.Sprintf("%s(%s)", labels["application"], labels["application_id"]))
	tempBody = strings.ReplaceAll(tempBody, "$application_id", labels["application_id"])
	tempBody = strings.ReplaceAll(tempBody, "$application_name", labels["application"])

	tempBody = strings.ReplaceAll(tempBody, "$service", labels["service"])
	tempBody = strings.ReplaceAll(tempBody, "$service_id", labels["service"])
	tempBody = strings.ReplaceAll(tempBody, "$service_name", labels["service"])
	tempBody = strings.ReplaceAll(tempBody, "ip", labels["ip"])

	return tempBody
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
			key:   v.Key,
			value: v.Value,
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
		metric:             metrics.Parse([]string{conf.Rule.Metric}),
		fuseConditionCount: conf.Rule.FuseCondition.Count,

		fuseTime: fuseTimeConf{
			time:    time.Duration(conf.Rule.FuseTime.Time) * time.Second,
			maxTime: time.Duration(conf.Rule.FuseTime.MaxTime) * time.Second,
		},
		recoverConditionCount: conf.Rule.RecoverCondition.Count,

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
