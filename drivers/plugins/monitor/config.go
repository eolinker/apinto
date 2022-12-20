package monitor

import "github.com/eolinker/eosc"

type Config struct {
	Output []eosc.RequireId `json:"output" skill:"github.com/eolinker/apinto/monitor-entry.monitor-entry.IOutput" label:"输出器列表"`
}
