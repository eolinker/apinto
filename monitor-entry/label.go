package monitor_entry

var (
	LabelApi      = "api"
	LabelApp      = "app"
	LabelUpstream = "upstream"
	LabelHandler  = "handler"
)

var labels = map[string]string{
	LabelApi:      "api",
	LabelApp:      "application",
	LabelHandler:  "handler",
	LabelUpstream: "service",
}
