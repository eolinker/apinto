package rate_limiting

type Config struct {
	Second           int64  `json:"second,omitempty"` // 每秒请求次数限制
	Minute           int64  `json:"minute,omitempty"` // 每分钟请求次数限制
	Hour             int64  `json:"hour,omitempty"`   // 每小时请求次数限制
	Day              int64  `json:"day,omitempty"`    // 每天请求次数限制
	HideClientHeader bool   `json:"hideClientHeader"` // 请求结果是否隐藏流控信息
	ResponseType     string `json:"responseType"`	  // 插件返回报错的类型
}


