package kubernetes

import (
	"context"
	"os"
	"testing"
)

func TestClient(t *testing.T) {
	cfg := &AccessConfig{
		Address:     []string{"https://172.18.189.43:6443"},
		Namespace:   "apinto",
		Inner:       false,
		BearerToken: os.Getenv("KUBERNETES_TOKEN"),
		Username:    "",
		Password:    "",
	}
	c, err := newClient(context.Background(), "apinto-gateway-stateful", cfg)
	if err != nil {
		t.Fatal(err)
	}
	list, err := c.GetNodeList("apinto-gateway")
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range list {
		t.Log(v)
	}
}
