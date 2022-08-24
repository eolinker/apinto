package jwt

import (
	"testing"
)

func TestJwtDecode(t *testing.T) {
	key := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJ1c2VyIjp7Im5hbWUiOiJhcGludG8ifX0.IqKYgppzwo75wGb3P_tBQeju-n-s8qpfjXPuE6Zyz9s"
	token, err := decodeToken(key)
	if err != nil {
		t.Error(err)
	}
	t.Log(token.Claims)
}
