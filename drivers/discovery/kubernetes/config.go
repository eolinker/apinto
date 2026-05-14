package kubernetes

// Config nacos驱动配置
type Config struct {
	Config *AccessConfig `json:"config" label:"配置信息"`
}

// AccessConfig 接入地址配置
type AccessConfig struct {
	Inner       bool     `json:"inner" label:"是否内部接入"`
	Address     []string `json:"address" label:"接入地址"`
	Namespace   string   `json:"namespace" label:"命名空间"`
	Username    string   `json:"username" label:"用户名"`
	Password    string   `json:"password" label:"密码"`
	BearerToken string   `json:"breaker_token" label:"Bearer Token"`
	PortName    string   `json:"port_name" label:"端口名称"`
}
