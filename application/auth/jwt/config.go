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
	Config *Rule   `json:"config"`
	Users  []*User `json:"users"`
}

type User struct {
	application.User
	Pattern Pattern `json:"pattern"`
}

type Pattern struct {
	Username string `json:"username"`
}

func (u *User) Username() string {
	return u.Pattern.Username
}

type Rule struct {
	Iss               string   `json:"iss" `
	Secret            string   `json:"secret"`
	RsaPublicKey      string   `json:"rsa_public_key"`
	Algorithm         string   `json:"algorithm"`
	ClaimsToVerify    []string `json:"claims_to_verify"`
	SignatureIsBase64 bool     `json:"signature_is_base_64"`
	Path              string   `json:"path"`
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
