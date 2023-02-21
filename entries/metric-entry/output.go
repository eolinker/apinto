package metric_entry

import (
	"github.com/eolinker/eosc"
)

const Skill = "github.com/eolinker/apinto/metric-entry.metric-entry.IOutput"

type IOutput interface {
	Output(metrics []string, entry eosc.IMetricEntry)
}
