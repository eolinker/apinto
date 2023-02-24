package prometheus

import "github.com/eolinker/eosc"

type Config struct {
	Output  []eosc.RequireId `json:"output" skill:"github.com/eolinker/apinto/metric-entry.metric-entry.IMetrics" label:"prometheus Output列表"`
	Metrics []string         `json:"metrics" yaml:"metrics" label:"指标名列表"`
}

const (
	globalScopeName = "prometheus"
)
