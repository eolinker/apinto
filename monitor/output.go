package monitor

const Skill = "github.com/eolinker/apinto/monitor.monitor.IOutput"

type IOutput interface {
	Output(point IPoint)
}
