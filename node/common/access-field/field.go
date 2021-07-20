package access_field

import (
	"reflect"
	"time"
)

type Proxy struct {
	Driver         string `json:"driver" field:"driver" desc:"请求类型驱动,目前只有http, 以后可能会有 mysql、grpc等,driver不一样,下面的字段不一样"`
	Request        string `json:"request" field:"request" desc:"请求信息, 如 POST https"`
	Uri            string `json:"uri" field:"uri" desc:"转发uri, /test"`
	Method         string `json:"method" field:"method" desc:"请求方法 如 POST"`
	Protocol       string `json:"protocol" field:"protocol" desc:"请求协议,http/https"`
	Upstream       string `json:"upstream" field:"upstream" desc:"负载信息"`
	Host           string `json:"host" field:"host" desc:"如果是upstream负载,这里是最终目标的ip/域名,否则这里应该根upstream一致"`
	RequestMsg     string `json:"request_msg" field:"request_msg" desc:"请求内容"`
	ResponseMsg    string `json:"response_msg" field:"response_msg" desc:"响应内容"`
	Status         int    `json:"status" field:"status" desc:"响应状态码"`
	RequestHeader  string `json:"request_header" field:"request_header" desc:"请求的header内容,格式为 raw"`
	ResponseHeader string `json:"response_header" field:"response_header" desc:"响应的header内容"`
	Timestamp      int64  `json:"timestamp" field:"timestamp" desc:"开始时间"`
	Timing         int64  `json:"timing" field:"timing" desc:"耗时" `
}

const (
	//DefaultTimeStampFormatter 时间戳默认格式化字符串
	DefaultTimeStampFormatter = "2006-01-02 15:04:05"
	//TimeIso8601Formatter iso8601格式化
	TimeIso8601Formatter = time.RFC3339
)

type Fields struct {
	RequestID         string   `json:"request_id" field:"request_id" desc:"请求唯一id"`
	Msec              int64    `json:"msec" field:"msec" desc:"日志写入时间。单位为秒，精度是毫秒。"`
	TimeLocal         string   `json:"time_local" field:"time_local" desc:"日志写入时间通用日志格式下的本地时间。"`
	TimeIso8601       string   `json:"time_iso8601" field:"time_iso8601" desc:"日志写入时间,ISO8601标准格式下的本地时间。"`
	Timestamp         int64    `json:"timestamp" field:"timestamp" desc:"请求时间戳/ms"`
	Timing            int64    `json:"timing" field:"timing" desc:"请求消耗时间/ms"`
	Service           string   `json:"service" field:"service" desc:"服务唯一id"`
	ServiceTitle      string   `json:"service_title" field:"service_title" desc:"服务名称（项目名称）"`
	Version           string   `json:"version" field:"version" desc:"服务版本"`
	Api               string   `json:"api" field:"api" desc:"api标示"`
	ApiTitle          string   `json:"api_title" field:"api_title" desc:"api标题,显示名,中文名"`
	ApiPath           string   `json:"api_path" field:"api_path" desc:"api监听的path"`
	Host              string   `json:"host" field:"host" desc:"用户请求的host"`
	GatewayIp         string   `json:"gateway_ip" field:"gateway_ip" desc:"网关节点信息ip,"`
	Cluster           string   `json:"cluster" field:"cluster" desc:"集群唯一id"`
	ClusterName       string   `json:"cluster_name" field:"cluster_name" desc:"集群名称"`
	StatusCode        int      `json:"status_code" field:"status_code" desc:"最终响应给前端的状态码"`
	RequestUri        string   `json:"request_uri" field:"request_uri" desc:"实际请求uri"`
	RequestMethod     string   `json:"request_method" field:"request_method" desc:"请求报文"`
	RequestMsg        string   `json:"request_msg" field:"request_msg" desc:"请求报文"`
	RequestMsgSize    int      `json:"request_msg_size" field:"request_msg_size" desc:"请求报文大小/kb"`
	RequestHeader     string   `json:"request_header" field:"request_header" desc:"请求中的header"`
	ResponseMsg       string   `json:"response_msg" field:"response_msg" desc:"响应报文"`
	ResponseMsgSize   int      `json:"response_msg_size" field:"response_msg_size" desc:"响应报文大小/kb"`
	ResponseHeader    string   `json:"response_header" field:"response_header" desc:"响应的header内容"`
	Proxys            []*Proxy `json:"proxys" field:"proxys" desc:"转发信息"`
	RemoteAddr        string   `json:"remote_addr" field:"remote_addr" desc:"记录客户端IP地址"`
	HTTPXForwardedFor string   `json:"http_x_forwarded_for" field:"http_x_forwarded_for" desc:"记录客户端IP地址(反向)"`
	HTTPReferer       string   `json:"http_referer" field:"http_referer" desc:"记录从哪个页面链接访问过来的"`
	HTTPUserAgent     string   `json:"http_user_agent" field:"http_user_agent" desc:"记录客户端浏览器相关信息"`
	Append            map[string]interface{}
}

func NewFields() *Fields {
	return &Fields{
		Append: make(map[string]interface{}),
	}
}

func (f *Fields) ToMap() map[string]interface{} {
	v := reflect.ValueOf(f).Elem()
	t := v.Type()
	n := t.NumField()
	m := make(map[string]interface{}, n)
	for i := 0; i < n; i++ {
		structField := t.Field(i)
		fn := structField.Tag.Get("field")
		if fn != "" {
			fv := v.Field(i)
			if fv.IsValid() && !isEmpty(fv) {
				m[fn] = fv.Interface()
			}
		}
	}
	for k, v := range f.Append {
		m[k] = v
	}
	return m
}

func isEmpty(v reflect.Value) bool {
	k := v.Kind()
	switch k {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.UnsafePointer, reflect.Interface, reflect.Slice:
		return v.IsNil()
	case reflect.String:
		return v.String() == ""
	}
	return false
}
