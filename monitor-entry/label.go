package monitor_entry

var (
	LabelApi      = "api"
	LabelApp      = "app"
	LabelUpstream = "upstream"
)

var labels = map[string]string{
	LabelApi:      "api_id",
	LabelApp:      "application_id",
	LabelUpstream: "service_id",
}
