package utils

import (
	"fmt"
	"testing"
)

func TestAes(t *testing.T) {
	key := Md5("admin")
	enValue := AES_CBC_Encrypt([]byte(Md5("Key123qaz:admin")), []byte(key))
	deValue := AES_CBC_Decrypt(enValue, []byte(key))
	fmt.Println(enValue)
	fmt.Println(string(deValue))
}
