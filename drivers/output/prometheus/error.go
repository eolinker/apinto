package prometheus

import "errors"

var (
	errorConfigType  = errors.New("error config type")
	errorNullMetrics = errors.New("error metrics can't be null. ")
	errorNullMetric  = errors.New("error metric can't be null. ")

	errorNullLabelsFormat         = "error metric %s labels can't be null. "
	errorPathFormat               = `error path %s is illegal. `
	errorCollectorFormat          = `error collector %s is illegal. `
	errorMetricTypeFormat         = `error metric_type %s is illegal. `
	errorMetricReduplicatedFormat = `error metric %s is reduplicated. `
	errorLabelFormat              = `error label %s is illegal. `
	errorLabelReduplicatedFormat  = `error metric %s's label %s is reduplicated. `
)
