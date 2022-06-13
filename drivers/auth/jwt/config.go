package jwt

//Config JWT实例配置
type Config struct {
	Credentials       []JwtCredential `json:"credentials"`
	SignatureIsBase64 bool            `json:"signature_is_base64"`
	ClaimsToVerify    []string        `json:"claims_to_verify"`
	RunOnPreflight    bool            `json:"run_on_preflight"`
	HideCredentials   bool            `json:"hide_credentials"`
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
