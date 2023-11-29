package jwt

import (
	"errors"
	"strconv"
	"strings"

	"github.com/eolinker/apinto/application"
	"github.com/eolinker/apinto/utils"
)

type Config struct {
	application.Auth
	Config *Rule   `json:"config" label:"JWT配置"`
	Users  []*User `json:"users" label:"用户列表"`
}

type User struct {
	Pattern Pattern `json:"pattern" label:"用户信息"`
	application.User
}

type Pattern struct {
	Username string `json:"username" label:"用户名"`
}

func (u *User) Username() string {
	return u.Pattern.Username
}

type Rule struct {
	Iss               string   `json:"iss" label:"签发机构"`
	Algorithm         string   `json:"algorithm" label:"签名算法" enum:"HS256,HS384,HS512,RS256,RS384,RS512,ES256,ES384,ES512"`
	Secret            string   `json:"secret" label:"密钥" switch:"algorithm==='HS256'||algorithm==='HS384'||algorithm==='HS512'"`
	RsaPublicKey      string   `json:"rsa_public_key" label:"RSA公钥" switch:"algorithm!=='HS256'&&algorithm!=='HS384'&&algorithm!=='HS512'"`
	Path              string   `json:"path" label:"用户名JsonPath"`
	ClaimsToVerify    []string `json:"claims_to_verify" label:"检验字段"`
	SignatureIsBase64 bool     `json:"signature_is_base_64" label:"签名是否base64加密" switch:"algorithm==='HS256'||algorithm==='HS384'||algorithm==='HS512'"`
}

func (c *Rule) ToID() (string, error) {
	builder := strings.Builder{}
	switch c.Algorithm {
	case "HS256", "HS384", "HS512":
		builder.WriteString(strings.TrimSpace(c.Iss))
		builder.WriteString(strings.TrimSpace(c.Secret))
		builder.WriteString(strings.TrimSpace(c.Algorithm))
		builder.WriteString(strconv.FormatBool(c.SignatureIsBase64))
		builder.WriteString(strings.TrimSpace(c.Path))
		for _, claim := range c.ClaimsToVerify {
			builder.WriteString(strings.TrimSpace(claim))
		}

	case "RS256", "RS384", "RS512", "ES256", "ES384", "ES512":
		builder.WriteString(strings.TrimSpace(c.Iss))
		builder.WriteString(strings.TrimSpace(c.RsaPublicKey))
		builder.WriteString(strings.TrimSpace(c.Algorithm))
		builder.WriteString(strings.TrimSpace(c.Path))
		for _, claim := range c.ClaimsToVerify {
			builder.WriteString(strings.TrimSpace(claim))
		}
	default:
		return "", errors.New("invalid algorithm")
	}
	return utils.Md5(builder.String()), nil
}
