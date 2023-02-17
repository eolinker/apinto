package prometheus_entry

const Skill = "github.com/eolinker/apinto/prometheus-entry.prometheus-entry.IOutput"

type IOutput interface {
	Output(metrics []string, entry IPromEntry)
}
