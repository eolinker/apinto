package consul

//Config consul驱动配置
type Config struct {
	Name   string            `json:"name"`
	Driver string            `json:"driver"`
	Scheme string            `json:"scheme"`
	Labels map[string]string `json:"labels"`
	Config AccessConfig      `json:"config"`
}

//AccessConfig 接入地址配置
type AccessConfig struct {
	Address []string          `json:"address"`
	Params  map[string]string `json:"params"`
}
