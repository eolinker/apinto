package response_filter

type Config struct {
	BodyFilter   []string `json:"body_filter" label:"响应体过滤字段"`
	HeaderFilter []string `json:"header_filter" label:"响应头过滤字段"`
}
