package prometheus

import "errors"

var (
	errorConfigType  = errors.New("error config type")
	errorNullMetrics = errors.New("error metrics can't be null. ")

	errorNullLabelsFormat = "error metric %s labels can't be null. "
	errorPathFormat       = `error path %s is illegal. `
	errorMetricFormat     = `error metric %s is illegal. `
	errorLabelFormat      = `error label %s is illegal. `
)
