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
	Iss          string            `json:"iss" label:"证书签发者" description:"playload计算内容之一"`
	Secret       string            `json:"secret" label:"密钥" description:"加密算法是HS时必填，用于校验token"`
	RSAPublicKey string            `json:"rsa_public_key" label:"RSA公钥" description:"加密算法是RS或ES时必填，用于校验token"`
	Algorithm    string            `json:"algorithm" enum:"HS256,HS384,HS512,RS256,RS384,RS512,ES256,ES384,ES512" label:"签名算法"`
	Labels       map[string]string `json:"labels" label:"用户标签"`
}
