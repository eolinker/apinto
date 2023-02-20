package prometheus_entry

import (
	"github.com/eolinker/eosc"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
)

const Skill = "github.com/eolinker/apinto/prometheus-entry.prometheus-entry.IOutput"

type IOutput interface {
	Output(metrics []string, entry eosc.IEntry, ctx http_context.IHttpContext)
}
