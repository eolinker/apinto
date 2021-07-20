package eureka

//Config eureka驱动配置
type Config struct {
	Name   string `json:"name"`
	Driver string `json:"driver"`
	Labels map[string]string
	Config AccessConfig `json:"config"`
}

//AccessConfig 接入地址配置
type AccessConfig struct {
	Address []string
	Params  map[string]string
}
