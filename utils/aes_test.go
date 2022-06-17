package utils

import (
	"fmt"
	"testing"
)

func TestAes(t *testing.T) {
	key := Md5("open-api")
	enValue := AES_CBC_Encrypt([]byte(Md5("Key123qaz:open-api")), []byte(key))
	deValue := AES_CBC_Decrypt(enValue, []byte(key))
	fmt.Println(enValue)
	fmt.Println(string(deValue))
}
