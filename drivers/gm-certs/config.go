package gm_certs

type Config struct {
	Name     string `json:"name" label:"名称" description:"名称"`
	SignKey  string `json:"sign_key" label:"签名密钥内容" format:"file" description:"密钥文件的后缀名一般为.key"`
	SignCert string `json:"sign_cert" label:"签名凭证内容" format:"file" description:"证书文件的后缀名一般为.crt 或 .pem"`
	EncKey   string `json:"enc_key" label:"加密密钥内容" format:"file" description:"密钥文件的后缀名一般为.key"`
	EncCert  string `json:"enc_cert" label:"加密凭证内容" format:"file" description:"证书文件的后缀名一般为.crt 或 .pem"`
}
