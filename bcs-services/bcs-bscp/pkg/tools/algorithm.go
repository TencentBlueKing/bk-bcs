package tools

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"math/rand"
)

var (
	AES = "aes"
)

// CreateCredential create credential
func CreateCredential(masterKey, encryptionAlgorithm string) (string, error) {
	if len(masterKey) == 0 || len(encryptionAlgorithm) == 0 {
		return "", fmt.Errorf("key or encryption algorithm is null")
	}
	algorithmText := randStr(32)
	switch encryptionAlgorithm {
	case AES:
		return AesEncrypt([]byte(algorithmText), []byte(masterKey))
	default:
		return "", fmt.Errorf("algorithm type is is not supported, type: %s", encryptionAlgorithm)
	}
}

// DecryptCredential Decrypt credential
func DecryptCredential(credential, masterKey, encryptionAlgorithm string) (string, error) {
	if len(masterKey) == 0 || len(encryptionAlgorithm) == 0 {
		return "", fmt.Errorf("key or encryption algorithm is null")
	}
	b64Byte, _ := base64.StdEncoding.DecodeString(credential)
	switch encryptionAlgorithm {
	case AES:
		return AesDecrypt(b64Byte, []byte(masterKey))
	default:
		return "", fmt.Errorf("algorithm type is is not supported, type: %s", encryptionAlgorithm)
	}
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// randStr 随机生成字符串
func randStr(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// PKCS7Padding PKCS7Padding
func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// PKCS7UnPadding PKCS7UnPadding
func PKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

// AesEncrypt AES加密,CBC
func AesEncrypt(origData, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	blockSize := block.BlockSize()
	origData = PKCS7Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return base64.StdEncoding.EncodeToString(crypted), nil
}

// AesDecrypt AES解密
func AesDecrypt(crypted, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = PKCS7UnPadding(origData)
	return string(origData), nil
}
