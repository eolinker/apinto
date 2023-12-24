package rsa_filter

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
)

var defaultPssOption = &rsa.PSSOptions{
	SaltLength: 32,
	Hash:       crypto.SHA256,
}

func encrypt(data []byte, publicKey *rsa.PublicKey) ([]byte, error) {
	blockSize := publicKey.Size() - 66
	var encryptedChunks [][]byte
	for len(data) > 0 {
		chunkSize := len(data)
		if chunkSize > blockSize {
			chunkSize = blockSize
		}
		// 取出要加密的块
		chunk := data[:chunkSize]
		data = data[chunkSize:]

		// 使用RSA公钥加密块
		encryptedChunk, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, chunk, nil)
		if err != nil {
			return nil, err
		}

		encryptedChunks = append(encryptedChunks, encryptedChunk)
	}
	return bytes.Join(encryptedChunks, nil), nil
}

func decrypt(data []byte, privateKey *rsa.PrivateKey) ([]byte, error) {

	blockSize := (privateKey.N.BitLen() + 7) / 8
	var decryptedChunks [][]byte
	for len(data) > 0 {
		chunkSize := len(data)
		if chunkSize > blockSize {
			chunkSize = blockSize
		}
		// 取出要加密的块
		chunk := data[:chunkSize]
		data = data[chunkSize:]

		// 使用RSA公钥加密块
		encryptedChunk, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, chunk, nil)
		if err != nil {
			return nil, err
		}

		decryptedChunks = append(decryptedChunks, encryptedChunk)
	}
	return bytes.Join(decryptedChunks, nil), nil
}

func sign(data []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	hashed := sha256.Sum256(data)
	return rsa.SignPSS(rand.Reader, privateKey, crypto.SHA256, hashed[:], defaultPssOption)
}

func verify(data []byte, signature []byte, publicKey *rsa.PublicKey) error {
	hashed := sha256.Sum256(data)
	return rsa.VerifyPSS(publicKey, crypto.SHA256, hashed[:], signature, defaultPssOption)
}

func parsePrivateKey(data []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("private key error")
	}
	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key err")
	}
	// 转换为 RSA 公钥
	rsaPrivateKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key is not an RSA public key")
	}
	return rsaPrivateKey, nil
}

func parsePublicKey(data []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("private key error")
	}
	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key err: %w", err)
	}
	// 转换为 RSA 公钥
	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not an RSA public key")
	}
	return rsaPublicKey, nil
}
