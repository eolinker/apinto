package prometheus

type iMetric interface {
	Set(value float64, labels map[string]string)
}
