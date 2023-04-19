package http_mocking

type Config struct {
	ResponseStatus  int               `json:"response_status" default:"200" label:"响应状态码" description:"仅http路由有效"`
	ContentType     string            `json:"content_type" label:"响应 Content-Type" enum:"application/json"`
	ResponseExample string            `json:"response_example"  format:"text" label:"响应Body" description:"与Json Schema字段二选一"`
	ResponseSchema  string            `json:"response_schema" format:"text" label:"Json Schema" description:"Mock生成的Json Schema语法数据，与响应Body字段二选一"`
	ResponseHeader  map[string]string `json:"response_header" label:"响应头"`
}

const (
	contentTypeJson = "application/json"
)
