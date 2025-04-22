package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

// AESCipher 封装AES加密解密功能
type AESCipher struct {
	key []byte
}

// NewAESCipher 初始化AES加密器
// key: 输入的密钥（任意长度）
// keyLength: 目标密钥长度（16=AES-128, 24=AES-192, 32=AES-256）
func NewAESCipher(key string, salt string, keyLength int) (*AESCipher, error) {
	if keyLength != 16 && keyLength != 24 && keyLength != 32 {
		return nil, errors.New("key length must be 16 (AES-128), 24 (AES-192), or 32 (AES-256) bytes")
	}

	keyBytes := []byte(key)
	var finalKey []byte

	switch {
	case len(keyBytes) == keyLength:
		// 密钥长度匹配，直接使用
		finalKey = keyBytes
	case len(keyBytes) < keyLength:
		// 密钥过短，使用PBKDF2扩展
		iterations := 100000
		finalKey = pbkdf2.Key(keyBytes, []byte(salt), iterations, keyLength, sha256.New)
	default:
		// 密钥过长，使用SHA-256哈希并截断
		hash := sha256.Sum256(keyBytes)
		finalKey = hash[:keyLength] // 截断到目标长度（16、24或32）
	}

	return &AESCipher{key: finalKey}, nil
}

// padData 为数据填充到AES块大小（16字节），使用PKCS#7填充
func padData(data []byte) []byte {
	padding := aes.BlockSize - len(data)%aes.BlockSize
	padText := make([]byte, len(data)+padding)
	copy(padText, data)
	for i := len(data); i < len(padText); i++ {
		padText[i] = byte(padding)
	}
	return padText
}

// unpadData 去除PKCS#7填充
func unpadData(data []byte) ([]byte, error) {
	if len(data) == 0 || len(data)%aes.BlockSize != 0 {
		return nil, errors.New("invalid padding")
	}
	padding := int(data[len(data)-1])
	if padding > aes.BlockSize || padding == 0 {
		return nil, errors.New("invalid padding")
	}
	for i := len(data) - padding; i < len(data); i++ {
		if data[i] != byte(padding) {
			return nil, errors.New("invalid padding")
		}
	}
	return data[:len(data)-padding], nil
}

// EncryptECB ECB模式加密
func (c *AESCipher) EncryptECB(plaintext string) (string, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", err
	}
	data := padData([]byte(plaintext))
	if len(data)%aes.BlockSize != 0 {
		return "", errors.New("plaintext is not a multiple of the block size")
	}
	ciphertext := make([]byte, len(data))
	for i := 0; i < len(data); i += aes.BlockSize {
		block.Encrypt(ciphertext[i:i+aes.BlockSize], data[i:i+aes.BlockSize])
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptECB ECB模式解密
func (c *AESCipher) DecryptECB(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	if len(data)%aes.BlockSize != 0 {
		return "", errors.New("ciphertext is not a multiple of the block size")
	}
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", err
	}
	plaintext := make([]byte, len(data))
	for i := 0; i < len(data); i += aes.BlockSize {
		block.Decrypt(plaintext[i:i+aes.BlockSize], data[i:i+aes.BlockSize])
	}
	unpadded, err := unpadData(plaintext)
	if err != nil {
		return "", err
	}
	return string(unpadded), nil
}

// EncryptCBC CBC模式加密
func (c *AESCipher) EncryptCBC(plaintext string) (string, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", err
	}
	data := padData([]byte(plaintext))
	ciphertext := make([]byte, aes.BlockSize+len(data))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], data)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptCBC CBC模式解密
func (c *AESCipher) DecryptCBC(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	if len(data) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}
	iv := data[:aes.BlockSize]
	data = data[aes.BlockSize:]
	if len(data)%aes.BlockSize != 0 {
		return "", errors.New("ciphertext is not a multiple of the block size")
	}
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", err
	}
	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(data))
	mode.CryptBlocks(plaintext, data)
	unpadded, err := unpadData(plaintext)
	if err != nil {
		return "", err
	}
	return string(unpadded), nil
}

// EncryptCFB CFB模式加密
func (c *AESCipher) EncryptCFB(plaintext string) (string, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", err
	}
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], []byte(plaintext))
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptCFB CFB模式解密
func (c *AESCipher) DecryptCFB(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	if len(data) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}
	iv := data[:aes.BlockSize]
	data = data[aes.BlockSize:]
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", err
	}
	stream := cipher.NewCFBDecrypter(block, iv)
	plaintext := make([]byte, len(data))
	stream.XORKeyStream(plaintext, data)
	return string(plaintext), nil
}

// EncryptOFB OFB模式加密
func (c *AESCipher) EncryptOFB(plaintext string) (string, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", err
	}
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}
	stream := cipher.NewOFB(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], []byte(plaintext))
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptOFB OFB模式解密
func (c *AESCipher) DecryptOFB(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	if len(data) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}
	iv := data[:aes.BlockSize]
	data = data[aes.BlockSize:]
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", err
	}
	stream := cipher.NewOFB(block, iv)
	plaintext := make([]byte, len(data))
	stream.XORKeyStream(plaintext, data)
	return string(plaintext), nil
}

// EncryptCTR CTR模式加密
func (c *AESCipher) EncryptCTR(plaintext string) (string, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", err
	}
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	nonce := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	stream := cipher.NewCTR(block, nonce)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], []byte(plaintext))
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptCTR CTR模式解密
func (c *AESCipher) DecryptCTR(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	if len(data) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}
	nonce := data[:aes.BlockSize]
	data = data[aes.BlockSize:]
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", err
	}
	stream := cipher.NewCTR(block, nonce)
	plaintext := make([]byte, len(data))
	stream.XORKeyStream(plaintext, data)
	return string(plaintext), nil
}

//// 示例使用
//func main() {
//	// 测试AES-256
//	cipher, err := NewAESCipher(32)
//	if err != nil {
//		panic(err)
//	}
//
//	plaintext := "Hello, AES Encryption!"
//
//	// 测试ECB
//	ecbEnc, err := cipher.EncryptECB(plaintext)
//	if err != nil {
//		panic(err)
//	}
//	ecbDec, err := cipher.DecryptECB(ecbEnc)
//	if err != nil {
//		panic(err)
//	}
//	println("ECB:", ecbDec)
//
//	// 测试CBC
//	cbcEnc, err := cipher.EncryptCBC(plaintext)
//	if err != nil {
//		panic(err)
//	}
//	cbcDec, err := cipher.DecryptCBC(cbcEnc)
//	if err != nil {
//		panic(err)
//	}
//	println("CBC:", cbcDec)
//
//	// 测试CFB
//	cfbEnc, err := cipher.EncryptCFB(plaintext)
//	if err != nil {
//		panic(err)
//	}
//	cfbDec, err := cipher.DecryptCFB(cfbEnc)
//	if err != nil {
//		panic(err)
//	}
//	println("CFB:", cfbDec)
//
//	// 测试OFB
//	ofbEnc, err := cipher.EncryptOFB(plaintext)
//	if err != nil {
//		panic(err)
//	}
//	ofbDec, err := cipher.DecryptOFB(ofbEnc)
//	if err != nil {
//		panic(err)
//	}
//	println("OFB:", ofbDec)
//
//	// 测试CTR
//	ctrEnc, err := cipher.EncryptCTR(plaintext)
//	if err != nil {
//		panic(err)
//	}
//	ctrDec, err := cipher.DecryptCTR(ctrEnc)
//	if err != nil {
//		panic(err)
//	}
//	println("CTR:", ctrDec)
//}
