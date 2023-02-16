package prometheus

type Config struct {
	Scopes  []string       `json:"scopes" label:"作用域"`
	Path    string         `json:"path" yaml:"path" label:"请求路径"`
	Metrics []MetricConfig `json:"metrics" yaml:"metrics" label:"指标列表"`
}

// TODO Labels枚举根据Metric是请求指标还是转发指标，显示的enum需要不一样
type MetricConfig struct {
	Metric string   `json:"metric" yaml:"metric" label:"指标名" enum:"request_total,request_timing,request_retry,request_req,request_resp,proxy_total,proxy_timing,proxy_req,proxy_resp"`
	Labels []string `json:"labels" yaml:"labels" label:"标签列表" enum:"node,cluster,method,upstream,status,api,app,host,handler,addr,path"`
}

type MetricType string

const (
	typeRequestMetric = "request"
	typeProxyMetric   = "proxy"
)

var (
	metricSet = map[string]string{
		"request_total":  typeRequestMetric,
		"request_timing": typeRequestMetric,
		"request_retry":  typeRequestMetric,
		"request_req":    typeRequestMetric,
		"request_resp":   typeRequestMetric,
		"proxy_total":    typeProxyMetric,
		"proxy_timing":   typeProxyMetric,
		"proxy_req":      typeProxyMetric,
		"proxy_resp":     typeProxyMetric,
	}
	metricLabelSet = map[string]map[string]struct{}{
		typeRequestMetric: {
			"node":     struct{}{},
			"cluster":  struct{}{},
			"method":   struct{}{},
			"upstream": struct{}{},
			"status":   struct{}{},
			"api":      struct{}{},
			"app":      struct{}{},
			"host":     struct{}{},
			"handler":  struct{}{},
		},
		typeProxyMetric: {
			"node":     struct{}{},
			"cluster":  struct{}{},
			"method":   struct{}{},
			"upstream": struct{}{},
			"status":   struct{}{},
			"addr":     struct{}{},
			"path":     struct{}{},
		},
	}
)
