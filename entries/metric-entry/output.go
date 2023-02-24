package metric_entry

import (
	"github.com/eolinker/eosc"
)

const Skill = "github.com/eolinker/apinto/metric-entry.metric-entry.IMetrics"

type IMetrics interface {
	Collect(metrics []string, entry eosc.IMetricEntry)
}
