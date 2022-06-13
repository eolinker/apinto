package jwt

//Config JWT实例配置
type Config struct {
	Credentials       []JwtCredential `json:"credentials" label:"证书列表"`
	SignatureIsBase64 bool            `json:"signature_is_base64" label:"base64加密"`
	ClaimsToVerify    []string        `json:"claims_to_verify" label:"校验字段"`
	HideCredentials   bool            `json:"hide_credentials" label:"是否隐藏证书"`
}

type jwtUsers struct {
	credentials []JwtCredential
}

//JwtCredential JWT验证信息
type JwtCredential struct {
	Iss          string            `json:"iss"`
	Secret       string            `json:"secret"`
	RSAPublicKey string            `json:"rsa_public_key"`
	Algorithm    string            `json:"algorithm"`
	Labels       map[string]string `json:"labels"`
}
