package certs

type Config struct {
	Name string `json:"name" label:"证书名称" description:"证书名称"`
	Key  string `json:"key" label:"密钥内容" format:"file" description:"密钥文件的后缀名一般为.key"`
	Pem  string `json:"pem" label:"证书内容" format:"file" description:"证书文件的后缀名一般为.crt 或 .pem"`
}
