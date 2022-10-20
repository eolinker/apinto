package fuse_strategy

import (
	"context"
	"fmt"
	"github.com/coocood/freecache"
	"github.com/eolinker/apinto/metrics"
	"github.com/eolinker/apinto/resources"
	"github.com/eolinker/apinto/strategy"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
	"github.com/go-redis/redis/v8"
	"time"
)

type fuseStatus string

const (
	fuseStatusNone    fuseStatus = "none"    //默认状态
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

func (f *FuseHandler) Fusing(eoCtx eocontext.EoContext, cache resources.ICache) bool {
	httpCtx, _ := http_service.Assert(eoCtx)

	fuseCondition := f.rule.fuseCondition
	recoverCondition := f.rule.recoverCondition

	ctx := context.Background()
	statusCode := httpCtx.Response().StatusCode()

	for _, code := range fuseCondition.statusCodes {
		if statusCode != code {
			continue
		}
		tx := cache.Tx()
		//记录失败count
		countKey := f.getFuseCountKey(eoCtx)

		if f.status == fuseStatusFusing {
			//缓存中拿不到数据 表示key过期 也就是熔断期已过 变成观察期
			if _, err := tx.Get(ctx, f.getFuseTimeKey(eoCtx)).Bytes(); err != nil && (err == freecache.ErrNotFound || err == redis.Nil) {
				f.status = fuseStatusObserve
				_ = tx.Exec(ctx)
				return false
			}
		}

		result, err := tx.IncrBy(ctx, countKey, 1, time.Second).Result()
		if err != nil {
			log.Errorf("FuseHandler Fusing %v", err)
			_ = tx.Exec(ctx)
			return true
		}

		//清除恢复的计数器
		tx.Del(ctx, f.getRecoverCountKey(eoCtx))

		if result >= fuseCondition.count {
			surplus := result % fuseCondition.count
			if surplus == 0 {
				//熔断持续时间=连续熔断次数*持续时间
				exp := time.Second * time.Duration((result/fuseCondition.count)*f.rule.fuseTime.time)
				maxExp := time.Duration(f.rule.fuseTime.maxTime) * time.Second
				if exp >= maxExp {
					exp = maxExp
				}
				tx.Set(ctx, f.getFuseTimeKey(eoCtx), []byte(""), exp)
				f.status = fuseStatusFusing
			}
			_ = tx.Exec(ctx)
			return true
		}
		break
	}

	for _, code := range recoverCondition.statusCodes {
		if code != statusCode {
			continue
		}
		if f.status == fuseStatusObserve || f.status == fuseStatusFusing {
			tx := cache.Tx()
			result, err := tx.IncrBy(ctx, f.getRecoverCountKey(eoCtx), 1, time.Second).Result()
			if err != nil {
				_ = tx.Exec(ctx)
				log.Errorf("FuseHandler Fusing %v", err)
				return true
			}

			//恢复正常期
			if result == recoverCondition.count {
				f.status = fuseStatusHealthy
			}
			_ = tx.Exec(ctx)
		}
		break
	}

	if f.status == fuseStatusHealthy || f.status == fuseStatusNone || f.status == fuseStatusObserve {
		return false
	}

	return true
}

func (f *FuseHandler) getFuseCountKey(label metrics.LabelReader) string {
	return fmt.Sprintf("fuse_%s_%s_%d", f.name, f.rule.metric.Metrics(label), time.Now().Second())
}

func (f *FuseHandler) getFuseTimeKey(label metrics.LabelReader) string {
	return fmt.Sprintf("fuse_time_%s_%s", f.name, f.rule.metric.Metrics(label))
}

func (f *FuseHandler) getRecoverCountKey(label metrics.LabelReader) string {
	return fmt.Sprintf("fuse_recover_%s_%s_%d", f.name, f.rule.metric.Metrics(label), time.Now().Second())
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
