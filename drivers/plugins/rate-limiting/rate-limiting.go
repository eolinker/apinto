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
	rateHourType = "Hour"
	rateDayType = "Day"
)

type RateLimiting struct {
	*Driver
	id    string
	name  string
	responseType string
}

func (r *RateLimiting) DoFilter(ctx http_service.IHttpContext, next http_service.IChain) (err error) {

	if next != nil {
		return next.DoChain(ctx)
	}
	return nil
}

func (r *RateLimiting) Id() string {
	return r.id
}

func (r *RateLimiting) Start() error {
	return nil
}

func (r *RateLimiting) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
	panic("implement me")
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

func (r *RateLimiting) addRateHeader(limit int64, remain int64, rateType string, ctx http_service.IHttpContext) {
	if limit == 0 || remain == 0 {
		return
	}
	ctx.Set().AddHeader(fmt.Sprintf("X-RateLimit-Limit-%s", rateType), strconv.FormatInt(limit, 10))
	ctx.Set().AddHeader(fmt.Sprintf("X-RateLimit-Remaining-%s", rateType), strconv.FormatInt(limit-remain, 10))
	return
}

