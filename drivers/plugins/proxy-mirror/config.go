package proxy_mirror

type Config struct {
	Host       string        `json:"host" label:"服务地址" description:"镜像服务地址, 需要包含scheme"`
	Path       string        `json:"path" label:"请求路径" description:"镜像的请求路径, 不填则默认使用当前路径"`
	PathMode   string        `json:"path_mode" label:"请求路径拼接模式" enum:"replace,prefix"`
	SampleConf *SampleConfig `json:"sample_conf" label:"采样配置"`
	Timeout    int           `json:"timeout" label:"请求超时时间"`
}

type SampleConfig struct {
	RandomRange int `json:"random_range" label:"随机数范围"`
	RandomPivot int `json:"random_pivot" label:"随机数锚点"`
}

const (
	pathModeReplace = "replace"
	pathModePrefix  = "prefix"
)

func (c *Config) doCheck() error {
	//TODO

	//校验host

	//校验path

	//校验path_mode

	//校验采样配置

	//校验镜像请求超时时间

	return nil
}
