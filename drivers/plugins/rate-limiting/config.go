package rate_limiting

type Config struct {
	Second           int64  `json:"second,omitempty" label:"每秒请求次数限制"`           // 每秒请求次数限制
	Minute           int64  `json:"minute,omitempty" label:"每分钟请求次数限制"`          // 每分钟请求次数限制
	Hour             int64  `json:"hour,omitempty" label:"每小时请求次数限制"`            // 每小时请求次数限制
	Day              int64  `json:"day,omitempty" label:"每天请求次数限制"`              // 每天请求次数限制
	HideClientHeader bool   `json:"hide_client_header" label:"是否隐藏流控信息"`         // 请求结果是否隐藏流控信息
	ResponseType     string `json:"response_type" label:"报错格式" enum:"text,json"` // 插件返回报错的类型
}
