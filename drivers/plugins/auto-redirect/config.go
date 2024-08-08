package auto_redirect

type Config struct {
	MaxRedirectCount int    `json:"max_redirect_count" label:"最大重定向次数" description:"最大重定向次数"`
	PathPrefix       string `json:"path_prefix" label:"重定向前缀" description:"重定向前缀"`
	AutoRedirect     bool   `json:"auto_redirect" label:"是否自动重定向" description:"是否自动重定向"`
}

var ()

func (c *Config) doCheck() error {
	return nil
}
