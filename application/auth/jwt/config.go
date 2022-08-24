package jwt

import (
	"errors"
	"github.com/eolinker/apinto/utils"
	"strconv"
	"strings"
)

type Config struct {
	Iss               string   `json:"iss" mapstructure:"iss"`
	Secret            string   `json:"secret"`
	RsaPublicKey      string   `json:"rsa_public_key"`
	Algorithm         string   `json:"algorithm"`
	ClaimsToVerify    []string `json:"claims_to_verify"`
	SignatureIsBase64 bool     `json:"signature_is_base_64"`
	Path              string   `json:"path"`
}

func (c *Config) ToID() (string, error) {
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
