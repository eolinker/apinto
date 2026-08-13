package quota_limiting

import (
	context_label "github.com/eolinker/apinto/common/context-label"
	quota_limiting_strategy "github.com/eolinker/apinto/drivers/strategy/quota-limiting-strategy"
	eoscContext "github.com/eolinker/eosc/eocontext"
	"time"
)

func BuildQuotaKeyAndTTL(ctx eoscContext.EoContext, key context_label.IKeyGenerator, st quota_limiting_strategy.IStrategy, now time.Time) (string, time.Duration) {
	var ttl time.Duration
	
	return key.Key(ctx, func(ctx eoscContext.EoContext, label string) string {
		switch label {
		case "target_type":
			return st.TargetType()
		case "strategy":
			return st.Name()
		case "period":
			return st.Period().String()
		case "time_format":
			var timeStr string
			switch st.Period() {
			case quota_limiting_strategy.PeriodMinute:
				timeStr = now.Format("200601021504")
				ttl = time.Duration(60-now.Second())*time.Second + time.Minute
			case quota_limiting_strategy.PeriodHour:
				timeStr = now.Format("2006010215")
				ttl = time.Duration(3600-now.Minute()*60-now.Second())*time.Second + 10*time.Minute
			case quota_limiting_strategy.PeriodDay:
				timeStr = now.Format("20060102")
				ttl = time.Duration(86400-now.Hour()*3600-now.Minute()*60-now.Second())*time.Second + 20*time.Minute
			case quota_limiting_strategy.PeriodMonth:
				timeStr = now.Format("200601")
				nextMonth := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
				ttl = nextMonth.Sub(now) + 3600*time.Minute
			case quota_limiting_strategy.PeriodTotal:
				timeStr = "total"
				ttl = -1
			default:
				timeStr = now.Format("20060102150405")
				ttl = time.Minute
			}
			return timeStr
		}
		return ctx.GetLabel(label)
	}), ttl
}
