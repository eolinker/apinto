package google

import (
	"context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/dns/v2"
)

var (
	scopes = []string{
		dns.CloudPlatformReadOnlyScope,
		dns.CloudPlatformScope,
	}
)

func newToken(ctx context.Context, data []byte) (*oauth2.Token, error) {
	cfg, err := google.JWTConfigFromJSON(data, scopes...)
	if err != nil {
		return nil, err
	}
	return cfg.TokenSource(ctx).Token()
}
