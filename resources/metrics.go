package resources

type Metrics interface {
	Send(labels map[string]string, fields map[string]interface{})
}
