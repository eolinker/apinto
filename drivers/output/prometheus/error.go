package prometheus

import "errors"

var (
	errorConfigType      = errors.New("error config type")
	errorNullMetrics     = errors.New("error metrics can't be null. ")
	errorNullMetric      = errors.New("error metric can't be null. ")
	errorNullScopeMetric = errors.New("error scope can't be null string. ")

	errorNullLabelsFormat         = "error metric %s labels can't be null. "
	errorPathFormat               = `error path %s is illegal. `
	errorCollectorFormat          = `error collector %s is illegal. `
	errorMetricTypeFormat         = `error metric_type %s is illegal. `
	errorMetricReduplicatedFormat = `error metric %s is reduplicated. `
	errorLabelFormat              = `error label %s is illegal. `
	errorLabelReduplicatedFormat  = `error metric %s's label name %s is reduplicated. `
	errorNullLabelFormat          = `error metric %s's label can't be null string'. `
	errorObjectivesFormat         = `error metric %s's objectives %s is illegal. `
)
