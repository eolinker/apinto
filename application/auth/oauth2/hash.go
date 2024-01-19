package oauth2

import (
	"fmt"
	"strconv"
	"strings"
)

type hashRule struct {
	algorithm  string
	iterations int
	length     int
	salt       string
	value      string
}

func extractHashRule(hash string) (*hashRule, error) {
	parts := strings.Split(hash, "$")
	if len(parts) != 5 {
		return nil, fmt.Errorf("invalid hashed password format")
	}
	subParts := strings.Split(parts[2], ",")
	if len(subParts) != 2 {
		return nil, fmt.Errorf("invalid hashed sub part format")
	}
	iterationsIndex := strings.Index(subParts[0], "=")
	if iterationsIndex == -1 {
		return nil, fmt.Errorf("iterations not found")
	}
	iterations, err := strconv.Atoi(subParts[0][iterationsIndex+1:])
	if err != nil {
		return nil, fmt.Errorf("invalid iterations format")
	}
	lengthIndex := strings.Index(subParts[1], "=")
	if lengthIndex == -1 {
		return nil, fmt.Errorf("length not found")
	}
	length, err := strconv.Atoi(subParts[1][lengthIndex+1:])
	if err != nil {
		return nil, fmt.Errorf("invalid length format")
	}
	return &hashRule{
		algorithm:  parts[0],
		iterations: iterations,
		length:     length,
		salt:       parts[3],
		value:      parts[4],
	}, nil
}

//
//func hashSecret(secret []byte, saltLen int, iterations int, keyLength int) (string, error) {
//	if saltLen < 1 {
//		saltLen = 16
//	}
//	salt, err := generateRandomSalt(saltLen)
//	if err != nil {
//		return "", err
//	}
//	// 迭代次数和密钥长度
//	if iterations < 1 {
//		iterations = 10000
//	}
//	if keyLength < 1 {
//		keyLength = 32
//	}
//
//	// 使用 PBKDF2 密钥派生函数
//	key := pbkdf2.Key(secret, salt, iterations, keyLength, sha512.New)
//	return fmt.Sprintf("$pbkdf2-sha512$i=%d,l=%d$%s$%s", iterations, keyLength, base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(key)), nil
//}

//func generateRandomSalt(length int) ([]byte, error) {
//	// Create a byte slice with the specified length
//	salt := make([]byte, length)
//
//	// Use crypto/rand to fill the slice with random bytes
//	_, err := rand.Read(salt)
//	if err != nil {
//		return nil, err
//	}
//
//	// Return the salt as a hexadecimal string
//	return salt, nil
//}
