package prometheus

import "errors"

var (
	errNullMetric = errors.New("Check config fail. metric can't be null. ")

	errNotImpEntryFormat = "%s:worker not implement IMetricOutput"
)
