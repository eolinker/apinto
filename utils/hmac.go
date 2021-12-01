package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

//HMacBySha256 HMacBySha256
func HMacBySha256(key, toSign string) string {
	hash := hmac.New(sha256.New, []byte(key)) // 创建对应的sha256哈希加密算法
	hash.Write([]byte(toSign))                // 写入加密数据

	return hex.EncodeToString(hash.Sum(nil))
}

//HexEncode HexEncode
func HexEncode(body []byte) string {
	h := sha256.New()
	h.Write(body)
	return hex.EncodeToString(h.Sum(nil))
}
