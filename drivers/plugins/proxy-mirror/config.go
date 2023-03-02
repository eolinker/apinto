package proxy_mirror

type Config struct {
	Addr       string        `json:"Addr" label:"服务地址" description:"镜像服务地址, 需要包含scheme"`
	SampleConf *SampleConfig `json:"sample_conf" label:"采样配置"`
	Timeout    int           `json:"timeout" label:"请求超时时间"`
	PassHost   string        `json:"pass_host" enum:"pass,node,rewrite" default:"pass" label:"转发域名" description:"请求发给上游时的 host 设置选型，pass:将客户端的 host 透传给上游，node:使用addr中配置的host，rewrite:使用下面指定的host值"`
	Host       string        `json:"upstream_host" label:"上游host" description:"指定上游请求的host，只有在 转发域名 配置为 rewrite 时有效" switch:"pass_host==='rewrite'"`
}

type SampleConfig struct {
	RandomRange int `json:"random_range" label:"随机数范围"`
	RandomPivot int `json:"random_pivot" label:"随机数锚点"`
}

const (
	modePass    = "pass"
	modeNode    = "node"
	modeRewrite = "rewrite"
)

func (c *Config) doCheck() error {
	//TODO

	//校验host

	//校验采样配置

	//校验镜像请求超时时间

	return nil
}
