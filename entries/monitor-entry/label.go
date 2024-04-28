package monitor_entry

var (
	LabelApi      = "api"
	LabelApp      = "app"
	LabelUpstream = "upstream"
	LabelHandler  = "handler"
	LabelProvider = "provider"
)

var labels = map[string]string{
	LabelApi:      "api",
	LabelApp:      "application",
	LabelHandler:  "handler",
	LabelUpstream: "service",
	LabelProvider: "provider",
}
