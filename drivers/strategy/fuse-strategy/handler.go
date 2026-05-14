package fuse_strategy

import (
	"context"
	_ "embed"
	"github.com/eolinker/apinto/resources"
	"github.com/eolinker/apinto/strategy"
	"github.com/eolinker/apinto/utils/response"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
	"github.com/eolinker/eosc/metrics"
	"strings"
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
const (
	fuseStateHealthy = "healthy"
	fuseStateFusing  = "fusing"
	fuseStateObserve = "observe"

	hashFieldState     = "state"
	hashFieldExpireAt  = "expire_at" // unix second
	hashFieldFuseCount = "fuse_count"
	hashFieldErrCount  = "err_count"  // 短窗口错误计数
	hashFieldSuccCount = "succ_count" // 半开连续成功计数
)

var (
	//go:embed lua/check_fuse_status.lua
	checkFuseStatusLua string
	//go:embed lua/on_failure.lua
	onFailureLua string
	//go:embed lua/on_success.lua
	onSuccessLua string
)

func getFuseBaseKey(metrics string) string {
	// 统一使用 hash tag，避免 CROSSSLOT
	// metrics 建议提前 TrimSpace + ToLower 处理
	tag := "fuse:" + strings.ToLower(strings.TrimSpace(metrics))
	return "{" + tag + "}"
}

type FuseHandler struct {
	name     string
	filter   strategy.IFilter
	priority int
	stop     bool
	rule     *ruleHandler
}

func (f *FuseHandler) Do(httpCtx http_service.IHttpContext, cache resources.ICache, metrics string) {
	statusCode := httpCtx.Response().StatusCode()
	ctx := httpCtx.Context()
	rule := f.rule

	baseKey := getFuseBaseKey(metrics)
	nowUnix := time.Now().Unix()

	switch rule.codeStatusMap[statusCode] {
	case codeStatusError:
		res, err := cache.Run(ctx, onFailureLua, []string{baseKey},
			rule.fuseConditionCount,
			rule.fuseTime.time/time.Second,
			rule.fuseTime.maxTime/time.Second,
			nowUnix,
		).Result()
		if err != nil {
			log.Warnf("fuse on_failure lua error: %v", err)
			return
		}

		ret := res.([]interface{})
		action := ret[0].(string)

		if action == "trigger_fusing" || action == "ignored_in_fusing" {
			if action == "trigger_fusing" {
				expSec := ret[1].(int64)
				log.Infof("熔断触发/回退: metrics=%s, 持续 %d 秒", metrics, expSec)
			}
			f.rule.response.Response(httpCtx)
			httpCtx.WithValue("is_block", true)
			httpCtx.SetLabel("block_name", f.name)
			httpCtx.SetLabel("handler", "fuse")
			httpCtx.Response().SetHeader("Strategy-Fuse", f.name)
			return
		}

	case codeStatusSuccess:
		// 只在 observe 期才处理恢复逻辑
		currentState := checkFuseStatus(ctx, metrics, cache) // 可选：提前判断
		if currentState != fuseStateObserve {
			return
		}

		res, err := cache.Run(ctx, onSuccessLua, []string{baseKey},
			rule.recoverConditionCount,
		).Result()
		if err != nil {
			log.Warnf("fuse on_success lua error: %v", err)
			return
		}

		ret := res.([]interface{})
		action := ret[0].(string)

		if action == "recovered_to_healthy" {
			succCnt := ret[1].(int64)
			log.Infof("熔断恢复成功: metrics=%s, 连续成功 %d 次", metrics, succCnt)
		}
	}

	return
}

//// 统一提取 tag 部分（最稳）
//func fuseTag(metrics string) string {
//	return "fuse:" + metrics // 或 "fuse-metrics:" + metrics 避免冲突
//}
//
//// 所有 key 必须严格这样写
//func getErrorCountKey(metrics string) string {
//	tag := fuseTag(metrics)
//	return "{" + tag + "}:err" // 或 "{fuse:" + metrics + "}:err"
//}
//
//func getSuccessCountKey(metrics string) string {
//	return "{" + fuseTag(metrics) + "}:succ"
//}
//
//func getFuseStatusKey(metrics string) string {
//	return "{" + fuseTag(metrics) + "}:status"
//}
//
//func getFuseCountKey(metrics string) string {
//	return "{" + fuseTag(metrics) + "}:count"
//}
//
//func getLockerKey(metrics string) string {
//	return "{" + fuseTag(metrics) + "}:lock"
//}

//// 熔断次数的key
//func getFuseCountKey(metrics string) string {
//	return fmt.Sprintf("strategy-fuse:count:%s_%d", metrics, time.Now().Unix())
//}
//
//// 失败次数的key
//func getErrorCountKey(metrics string) string {
//	return fmt.Sprintf("strategy-fuse:error_count:%s_%d", metrics, time.Now().Unix())
//}
//
//func getSuccessCountKey(metrics string) string {
//	return fmt.Sprintf("strategy-fuse:success_count:%s_%d", metrics, time.Now().Unix())
//}
//func getFuseStatusKey(metrics string) string {
//	return fmt.Sprintf("strategy-fuse:status:%s", metrics)
//}

func checkFuseStatus(ctx context.Context, metrics string, cache resources.ICache) fuseStatus {

	baseKey := getFuseBaseKey(metrics)
	result, err := cache.Run(ctx, checkFuseStatusLua, []string{baseKey}, time.Now().Unix()).Result()
	if err != nil {
		log.Errorf("checkFuseStatus lua error: %v", err)
		return fuseStatusHealthy
	}
	ret := result.([]interface{})
	if len(ret) > 0 {
		status, ok := ret[0].(string)
		if !ok {
			log.Errorf("checkFuseStatus lua error: value is not string: %v", ret[0])
			return fuseStatusHealthy
		}
		return fuseStatus(status)
	}
	return fuseStatusHealthy
}

//func getFuseStatus(ctx context.Context, metrics string, cache resources.ICache) fuseStatus {
//
//	key := getFuseStatusKey(metrics)
//
//	// 获取熔断结束时间
//	expUnixStr, err := cache.Get(ctx, key).Result()
//
//	// 1. 如果 Key 不存在 (redis.Nil)，说明是健康状态
//	if err != nil {
//		return fuseStatusHealthy
//	}
//
//	// 2. 解析时间戳 (Lua 写入的是 base 10 字符串)
//	expUnix, err := strconv.ParseInt(expUnixStr, 10, 64)
//	if err != nil {
//		// 数据损坏，防御性处理为健康
//		return fuseStatusHealthy
//	}
//	now := time.Now().Unix()
//	// 3. 判断时间
//	// 如果 当前时间 > 熔断结束时间，说明物理Key还没过期，但逻辑熔断已结束 -> 进入观察期
//	if now > expUnix {
//		return fuseStatusObserve
//	}
//
//	// 否则 -> 熔断中
//	return fuseStatusFusing
//}

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
		metric:             metrics.ParseArray([]string{conf.Rule.Metric}, "-"),
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
