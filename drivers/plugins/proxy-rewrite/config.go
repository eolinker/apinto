package proxy_rewrite

import "fmt"

type Config struct {
	Scheme   string            `json:"scheme"`
	URI      string            `json:"uri"`
	RegexURI []string          `json:"regex_uri"`
	Host     string            `json:"host"`
	Headers  map[string]string `json:"headers"`
}

func (c *Config) doCheck() error {
	if c.Scheme == "" {
		c.Scheme = "http"
	} else if c.Scheme != "http" && c.Scheme != "https" {
		return fmt.Errorf(schemeErrInfo, c.Scheme)
	}

	//RegexURI切片要么为空，要么只有两个值,第一个值为正则匹配值，第二个是用于替换的正则字符串
	lenRegURI := len(c.RegexURI)
	if lenRegURI != 0 && lenRegURI != 2 {
		return fmt.Errorf(regexpURIErrInfo, c.RegexURI)
	}

	return nil
}
