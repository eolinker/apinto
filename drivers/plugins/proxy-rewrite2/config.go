package proxy_rewrite2

import "fmt"

type Config struct {
	PathType    string            `json:"path_type" label:"转发路径重写类型" enum:"none,static,prefix,regex"`
	StaticPath  string            `json:"static_path" label:"静态转发路径" switch:"path_type==='static'"`
	PrefixPath  []*SPrefixPath    `json:"prefix_path" label:"转发路径前缀替换" switch:"path_type==='prefix'"`
	RegexPath   []*SRegexPath     `json:"regex_path" label:"转发路径正则替换" switch:"path_type==='regex'"`
	HostRewrite bool              `json:"host_rewrite" label:"是否重写host" `
	Host        string            `json:"host" label:"Host" switch:"host_rewrite===true"`
	Headers     map[string]string `json:"headers" label:"请求头部"`
}

type SPrefixPath struct {
	PrefixPathMatch   string `json:"prefix_path_match" label:"转发路径前缀匹配字符串"`
	PrefixPathReplace string `json:"prefix_path_replace" label:"转发路径前缀替换字符串"`
}

type SRegexPath struct {
	RegexPathMatch   string `json:"regex_path_match" label:"转发路径正则匹配表达式"`
	RegexPathReplace string `json:"regex_path_replace" label:"转发路径正则替换表达式"`
}

func (c *Config) doCheck() error {
	switch c.PathType {
	case typeStatic, typePrefix, typeRegex:
	default:
		c.PathType = typeNone
	}

	if c.HostRewrite && len(c.Host) == 0 {
		return fmt.Errorf(hostErrInfo)
	}

	return nil
}
