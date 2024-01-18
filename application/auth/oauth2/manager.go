package oauth2

import "github.com/eolinker/eosc"

func registerClient(clientId string, client *Client) {
	manager.clients.Set(clientId, client)
}

func removeClient(clientId string) {
	manager.clients.Del(clientId)
}

func getClient(clientId string) (*Client, bool) {
	return manager.clients.Get(clientId)
}

var manager = NewManager()

// Manager 管理oauth2配置
type Manager struct {
	clients eosc.Untyped[string, *Client]
}

func NewManager() *Manager {
	return &Manager{clients: eosc.BuildUntyped[string, *Client]()}
}

type Client struct {
	*Pattern
	// Expire 过期时间
	Expire   int64
	hashRule *hashRule
}
