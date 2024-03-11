package openid_connect_jwt

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/eolinker/eosc/log"

	"github.com/lestrrat-go/jwx/jwk"
)

var client = http.Client{
	Timeout: 5 * time.Second,
}

type IssuerConfig struct {
	ID            string             `json:"id"`
	Issuer        string             `json:"issuer"`
	Configuration *DiscoveryConfig   `json:"configuration"`
	Keys          []JWK              `json:"keys"`
	UpdateTime    time.Time          `json:"update_time"`
	JWKKeys       map[string]jwk.Key `json:"-"`
}

type DiscoveryConfig struct {
	TokenEndpoint                     string   `json:"token_endpoint"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
	JwksUri                           string   `json:"jwks_uri"`
	ResponseModesSupported            []string `json:"response_modes_supported"`
	SubjectTypesSupported             []string `json:"subject_types_supported"`
	IdTokenSigningAlgValuesSupported  []string `json:"id_token_signing_alg_values_supported"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	ScopesSupported                   []string `json:"scopes_supported"`
	Issuer                            string   `json:"issuer"`
	MicrosoftMultiRefreshToken        bool     `json:"microsoft_multi_refresh_token"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	DeviceAuthorizationEndpoint       string   `json:"device_authorization_endpoint"`
	HttpLogoutSupported               bool     `json:"http_logout_supported"`
	FrontchannelLogoutSupported       bool     `json:"frontchannel_logout_supported"`
	EndSessionEndpoint                string   `json:"end_session_endpoint"`
	ClaimsSupported                   []string `json:"claims_supported"`
	CheckSessionIframe                string   `json:"check_session_iframe"`
	UserinfoEndpoint                  string   `json:"userinfo_endpoint"`
	KerberosEndpoint                  string   `json:"kerberos_endpoint"`
	TenantRegionScope                 string   `json:"tenant_region_scope"`
	CloudInstanceName                 string   `json:"cloud_instance_name"`
	CloudGraphHostName                string   `json:"cloud_graph_host_name"`
	MsgraphHost                       string   `json:"msgraph_host"`
	RbacUrl                           string   `json:"rbac_url"`
}

type JWKs struct {
	Keys []JWK `json:"keys"`
}

type JWK struct {
	Kid     string   `json:"kid"`
	Kty     string   `json:"kty"`
	Alg     string   `json:"alg"`
	Use     string   `json:"use"`
	N       string   `json:"n"`
	E       string   `json:"e"`
	X5C     []string `json:"x5c"`
	X5T     string   `json:"x5t"`
	X5TS256 string   `json:"x5t#S256"`
}

func getIssuerConfig(issuer string) (*DiscoveryConfig, error) {
	resp, err := client.Get(issuer)
	if err != nil {
		return nil, fmt.Errorf("get issuer config error: %w, issuer: %s", err, issuer)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read issuer config error: %w, issuer: %s", err, issuer)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("get issuer config error: %d, issuer: %s, body: %s", resp.StatusCode, issuer, string(body))
	}
	var config DiscoveryConfig
	err = json.Unmarshal(body, &config)
	if err != nil {
		return nil, fmt.Errorf("unmarshal issuer config error: %w, issuer: %s, body: %s", err, issuer, string(body))
	}
	return &config, nil
}

func getJWKs(uri string) ([]JWK, map[string]jwk.Key, error) {
	resp, err := client.Get(uri)
	if err != nil {
		return nil, nil, fmt.Errorf("get issuer jwks error: %w, uri: %s", err, uri)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("read issuer jwks error: %w, uri: %s", err, uri)
	}
	if resp.StatusCode != 200 {
		return nil, nil, fmt.Errorf("get issuer jwks error: %d, uri: %s, body: %s", resp.StatusCode, uri, string(body))
	}
	var jwks JWKs
	err = json.Unmarshal(body, &jwks)
	if err != nil {
		return nil, nil, fmt.Errorf("unmarshal issuer jwks error: %w, uri: %s, body: %s", err, uri, string(body))
	}
	set, err := jwk.Parse(body)
	if err != nil {
		return nil, nil, fmt.Errorf("parse issuer jwks error: %w, uri: %s, body: %s", err, uri, string(body))
	}
	keys := make(map[string]jwk.Key)
	l := set.Len()
	for i := 0; i < l; i++ {
		key, success := set.Get(i)
		if !success {
			continue
		}
		if key.KeyUsage() != string(jwk.ForSignature) {
			continue
		}
		pubKey, err := key.PublicKey()
		if err != nil {
			log.Errorf("get public key error: %w, uri: %s, key: %s", err, uri, key.KeyID())
		}
		keys[key.KeyID()] = pubKey
	}
	return jwks.Keys, keys, nil
}
