package auto_redirect

import "testing"

func TestParseUrl(t *testing.T) {
	t.Log(insertPrefix("http://test:8982/aaa/aaa", "/prefix"))
}
