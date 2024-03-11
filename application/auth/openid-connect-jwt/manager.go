package openid_connect_jwt

import (
	"time"

	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/log"
)

var (
	manager = NewManager()
)

type Manager struct {
	Issuers eosc.Untyped[string, *IssuerConfig]
	Apps    eosc.Untyped[string, map[string]struct{}]
}

func NewManager() *Manager {
	m := &Manager{Issuers: eosc.BuildUntyped[string, *IssuerConfig](), Apps: eosc.BuildUntyped[string, map[string]struct{}]()}
	go m.doLoop()
	return m
}

func (m *Manager) doLoop() {
	ticket := time.NewTicker(10 * time.Second)
	defer ticket.Stop()
	for {
		select {
		case <-ticket.C:
			for _, issuer := range m.Issuers.All() {
				config, err := getIssuerConfig(issuer.Issuer)
				if err != nil {
					log.Error(err)
					continue
				}
				issuer.Configuration = config
				issuer.UpdateTime = time.Now()
				keys, jwks, err := getJWKs(config.JwksUri)
				if err != nil {
					log.Errorf("%w, issuer: %s", err, issuer.Issuer)
					continue
				}
				issuer.Keys = keys
				issuer.JWKKeys = jwks
			}
		}
	}
}

func (m *Manager) Set(id string, config *IssuerConfig) {
	m.Issuers.Set(id, config)
}

func (m *Manager) Del(id string) {
	m.Issuers.Del(id)
}

func (m *Manager) GetIssuerIDMap(appID string) map[string]struct{} {
	result := make(map[string]struct{})
	idMap, has := m.Apps.Get(appID)
	if !has {
		return result
	}
	for id := range idMap {
		result[id] = struct{}{}
	}
	return result
}

func (m *Manager) SetIssuerIDMap(appID string, issuerIDMap map[string]struct{}) {
	m.Apps.Set(appID, issuerIDMap)
}

func (m *Manager) DelIssuerIDMap(appID string) (map[string]struct{}, bool) {
	return m.Apps.Del(appID)
}
