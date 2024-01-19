package oauth2

import "testing"

func TestHash(t *testing.T) {
	data := "$pbkdf2-sha512$i=10000,l=32$7BGLyS03BLF+F+M01p7MBg$OTAR1PTJpXzCVBfRq3VcGXYlSeRD2IUEzk/RsRQwfwI"

	_, err := extractHashRule(data)
	if err != nil {
		t.Fatal(err)
	}
}
