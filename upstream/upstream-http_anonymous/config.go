package upstream_http_anonymous

//Config 匿名http-proxy驱动配置
type Config struct {
	id     string
	Name   string `json:"name"`
	Driver string `json:"driver"`
}
