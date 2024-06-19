package auto_redirect

type Config struct {
	MaxRedirectCount int `json:"max-redirect-count" label:"最大重定向次数" description:"最大重定向次数"`
}

var ()

func (c *Config) doCheck() error {
	return nil
}
