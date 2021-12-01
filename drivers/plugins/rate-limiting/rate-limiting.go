package rate_limiting

import (
	"encoding/json"
	"fmt"
	"github.com/eolinker/eosc"
	http_service "github.com/eolinker/eosc/http-service"
	"strconv"
)

const (
	rateSecondType = "Second"
	rateMinuteType = "Minute"
	rateHourType   = "Hour"
	rateDayType    = "Day"
)

type RateLimiting struct {
	*Driver
	id               string
	name             string
	rateInfo         *rateInfo
	hideClientHeader bool
	responseType     string
}



func (r *RateLimiting) doLimit() (bool, string, int) {
	info := r.rateInfo
	if info == nil {
		return true, "", 200
	}
	if info.second != nil {
		ok := info.second.check()
		if !ok {
			return false, "[rate_limiting] API rate limit of second exceeded", 429
		}
	}
	if info.minute != nil {
		ok := info.minute.check()
		if !ok {
			return false, "[rate_limiting] API rate limit of minute exceeded", 429
		}
	}
	if info.hour != nil {
		ok := info.hour.check()
		if !ok {
			return false, "[rate_limiting] API rate limit of hour exceeded", 429
		}
	}
	if info.day != nil {
		ok := info.day.check()
		if !ok {
			return false, "[rate_limiting] API rate limit of day exceeded", 429
		}
	}
	return true, "", 200
}

func (r *RateLimiting) Destroy() {
	r.responseType = ""
	r.rateInfo.close()
	r.rateInfo = nil
}

func (r *RateLimiting) DoFilter(ctx http_service.IHttpContext, next http_service.IChain) (err error) {
	// 前置处理
	flag, result, status := r.doLimit()
	if !flag {
		// 超过限制
		resp := ctx.Response()
		result = r.responseEncode(result, status)
		resp.SetStatus(403, "403")
		resp.SetBody([]byte(result))
		return err
	}
	// 后置处理
	if next != nil {
		err = next.DoChain(ctx)
	}
	if !r.hideClientHeader {
		r.addRateHeader(ctx, rateSecondType)
		r.addRateHeader(ctx, rateMinuteType)
		r.addRateHeader(ctx, rateHourType)
		r.addRateHeader(ctx, rateHourType)
	}
	return err
}

func (r *RateLimiting) Id() string {
	return r.id
}

func (r *RateLimiting) Start() error {
	return nil
}

func (r *RateLimiting) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
	confObj, err := r.check(conf)
	if err != nil {
		return err
	}
	r.rateInfo = CreateRateInfo(confObj)
	r.hideClientHeader = confObj.HideClientHeader
	r.responseType = confObj.ResponseType
	return nil
}

func (r *RateLimiting) Stop() error {
	return nil
}

func (r *RateLimiting) CheckSkill(skill string) bool {
	return http_service.FilterSkillName == skill
}

func (r *RateLimiting) responseEncode(origin string, statusCode int) string {
	if r.responseType == "json" {
		tmp := map[string]interface{}{
			"message":     origin,
			"status_code": statusCode,
		}
		newInfo, _ := json.Marshal(tmp)
		return string(newInfo)
	}
	return origin
}

func (r *RateLimiting) addRateHeader(ctx http_service.IHttpContext, rateType string) {
	var rate *rateTimer
	switch rateType {
	case rateSecondType:
		rate = r.rateInfo.second
	case rateMinuteType:
		rate = r.rateInfo.minute
	case rateHourType:
		rate = r.rateInfo.hour
	case rateDayType:
		rate = r.rateInfo.day
	}
	// 不限制
	if rate == nil || rate.limitCount == 0 || rate.requestCount == 0 {
		return
	}
	resp := ctx.Response()
	resp.SetHeader(fmt.Sprintf("X-RateLimit-Limit-%s", rateType), strconv.FormatInt(rate.limitCount, 10))
	resp.SetHeader(fmt.Sprintf("X-RateLimit-Remaining-%s", rateType), strconv.FormatInt(rate.limitCount-  rate.requestCount, 10))
	return
}
