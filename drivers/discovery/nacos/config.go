package nacos

const defaultScheme = "http"

// Config nacos驱动配置
type Config struct {
	Config AccessConfig `json:"config" label:"配置信息"`
}

// AccessConfig 接入地址配置
type AccessConfig struct {
	Address []string          `json:"address" label:"nacos地址"`
	Params  map[string]string `json:"params" label:"参数"`
}
