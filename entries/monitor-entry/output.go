package monitor_entry

const Skill = "github.com/eolinker/apinto/monitor-entry.monitor-entry.IOutput"

type IOutput interface {
	Output(point ...IPoint)
}
