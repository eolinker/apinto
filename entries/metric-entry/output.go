package metric_entry

import (
	"github.com/eolinker/eosc"
)

const Skill = "github.com/eolinker/apinto/prometheus-entry.prometheus-entry.IOutput"

type IOutput interface {
	Output(metrics []string, entry eosc.IMetricEntry)
}
