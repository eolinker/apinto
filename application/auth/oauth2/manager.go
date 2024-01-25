package oauth2

import (
	"crypto/sha512"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/pbkdf2"

	"github.com/eolinker/eosc"
)

type IClient interface {
	ClientID() string
	ClientSecret() string
	ClientType() string
	HashSecret() bool
	RedirectUrls() []string
	MatchSecret(secret string) error
	Expire() int64
}

func registerClient(clientId string, client IClient) {
	manager.clients.Set(clientId, client)
}

func RemoveClient(clientId string) {
	manager.clients.Del(clientId)
}

func GetClient(clientId string) (IClient, bool) {
	return manager.clients.Get(clientId)
}

var manager = NewManager()

// Manager 管理oauth2配置
type Manager struct {
	clients eosc.Untyped[string, IClient]
}

func NewManager() *Manager {
	return &Manager{clients: eosc.BuildUntyped[string, IClient]()}
}

type client struct {
	clientId     string
	clientSecret string
	clientType   string
	hashSecret   bool
	redirectUrls []string
	expire       int64
	hashRule     *hashRule
}

func (c *client) ClientID() string {
	return c.clientId
}

func (c *client) ClientSecret() string {
	return c.clientSecret
}

func (c *client) ClientType() string {
	return c.clientType
}

func (c *client) HashSecret() bool {
	return c.hashSecret
}

func (c *client) RedirectUrls() []string {
	return c.redirectUrls
}

func (c *client) Expire() int64 {
	return c.expire
}

func (c *client) MatchSecret(clientSecret string) error {
	orgSecret := c.clientSecret
	if c.hashSecret {
		salt, _ := base64.RawStdEncoding.DecodeString(c.hashRule.salt)
		secret := pbkdf2.Key([]byte(clientSecret), salt, c.hashRule.iterations, c.hashRule.length, sha512.New)
		clientSecret = base64.RawStdEncoding.EncodeToString(secret)
		orgSecret = c.hashRule.value
	}

	if orgSecret != clientSecret {
		return fmt.Errorf("fail to match secret,now: %s,hope: %s,client id is %s", clientSecret, orgSecret, c.clientId)
	}
	return nil
}
