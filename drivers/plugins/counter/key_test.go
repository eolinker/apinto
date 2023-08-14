package counter

import "testing"

func TestGenerateKey(t *testing.T) {
	key := newKeyGenerate("a:$b:$c:d")
	t.Log(key.format, key.variables)
}
