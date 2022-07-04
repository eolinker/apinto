package proxy_rewrite

import "fmt"

type Config struct {
	Scheme   string            `json:"scheme" label:"协议"`
	URI      string            `json:"uri" label:"路径"`
	RegexURI []string          `json:"regex_uri" label:"正则替换路径（regex_uri）"`
	Host     string            `json:"host" label:"Host"`
	Headers  map[string]string `json:"headers" label:"请求头部"`
}

func (c *Config) doCheck() error {
	if c.Scheme == "" {
		c.Scheme = "http"
	} else if c.Scheme != "http" && c.Scheme != "https" {
		return fmt.Errorf(schemeErrInfo, c.Scheme)
	}

	lenRegURI := len(c.RegexURI)

	// URI和RegexURI至少选填其一
	if c.URI == "" && lenRegURI == 0 {
		return fmt.Errorf(uriErrInfo)
	}

	//RegexURI切片要么为空，要么只有两个值,第一个值为正则匹配值，第二个是用于替换的正则字符串
	if lenRegURI > 0 && lenRegURI != 2 {
		return fmt.Errorf(regexpURIErrInfo, c.RegexURI)
	}

	return nil
}
