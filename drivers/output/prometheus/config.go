package prometheus

type Config struct {
	Scopes  []string        `json:"scopes" label:"作用域"`
	Path    string          `json:"path" yaml:"path" label:"Metrics路径"`
	Metrics []*MetricConfig `json:"metrics" yaml:"metrics" label:"指标列表"`
}

// TODO Labels枚举根据Metric是请求指标还是转发指标，显示的enum需要不一样
type MetricConfig struct {
	Metric      string   `json:"metric" yaml:"metric" label:"指标名"`
	Description string   `json:"description" yaml:"description" label:"指标描述"`
	Collector   string   `json:"collector" yaml:"collector" label:"收集类型" enum:"request_total,request_timing,request_retry,request_req,request_resp,proxy_total,proxy_timing,proxy_req,proxy_resp"`
	Objectives  string   `json:"objectives" yaml:"objectives" label:"quantiles分位数值配置" description:"格式为0.5:0.01用,号隔开数值.当收集类型为request_timing,request_retry,request_req,request_resp,proxy_timing,proxy_req,proxy_resp时选填"`
	Labels      []string `json:"labels" yaml:"labels" label:"标签列表" description:"$表示引用变量,不带$表示使用常量,as表示使用别名. $node表示标签名为node，值使用node变量;$node as node_id表示标签名为node_id,值使用node变量;node as node_id表示常量，标签名为node_id,值为字符串node.变量可选:node,cluster,method,upstream,status,api,app,host,handler,addr,path"`
}

type MetricType string

const (
	typeRequestMetric = "request"
	typeProxyMetric   = "proxy"

	labelTypeVar   = "variable"
	labelTypeConst = "const"

	typeCounter   = "Counter"
	typeGauge     = "Gauge"
	typeHistogram = "Histogram"
	typeSummary   = "Summary"
)

var (
	collectorSet = map[string]string{
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

	//metricLabelSet = map[string]map[string]struct{}{
	//	typeRequestMetric: {
	//		"node":     {},
	//		"cluster":  {},
	//		"method":   {},
	//		"upstream": {},
	//		"status":   {},
	//		"api":      {},
	//		"app":      {},
	//		"host":     {},
	//		"handler":  {},
	//	},
	//	typeProxyMetric: {
	//		"node":     {},
	//		"cluster":  {},
	//		"method":   {},
	//		"upstream": {},
	//		"status":   {},
	//		"addr":     {},
	//		"path":     {},
	//	},
	//}
	//collectorTypeSet collector对应的指标类型
	collectorTypeSet = map[string]string{
		"request_total":  typeCounter,
		"request_timing": typeSummary,
		"request_retry":  typeSummary,
		"request_req":    typeSummary,
		"request_resp":   typeSummary,
		"proxy_total":    typeCounter,
		"proxy_timing":   typeSummary,
		"proxy_req":      typeSummary,
		"proxy_resp":     typeSummary,
	}
	metricTypeSet = map[string]struct{}{
		typeCounter:   {},
		typeGauge:     {},
		typeHistogram: {},
		typeSummary:   {},
	}
)
