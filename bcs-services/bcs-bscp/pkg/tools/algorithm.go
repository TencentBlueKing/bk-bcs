package tools

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"encoding/base64"
	"fmt"
	"math/rand"
)

var (
	AES = "aes"
	DES = "des"
)

// CreateCredential
func CreateCredential(bizId uint32, masterKey, encryptionAlgorithm string) (string, error) {
	if len(masterKey) == 0 || len(encryptionAlgorithm) == 0 {
		return "", fmt.Errorf("key or encryption algorithm is null")
	}
	algorithmText := fmt.Sprintf("%d-%s-%s", bizId, encryptionAlgorithm, randStr(10))
	switch encryptionAlgorithm {
	case AES:
		return AesEncrypt([]byte(algorithmText), []byte(masterKey))
	case DES:
		return DesEncryptToBase([]byte(algorithmText), []byte(masterKey))
	default:
		return "", fmt.Errorf("algorithm type is is not supported, type: %s", encryptionAlgorithm)
	}
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// randStr 随机生成字符串
func randStr(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// PKCS5Padding size padding
func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// PKCS5UnPadding size unpadding
func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

// DesEncryptToBase encrypt with priKey simply, out base64 string
func DesEncryptToBase(origData, key []byte) (string, error) {
	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		return "", err
	}
	src := PKCS5Padding(origData, block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, key[:block.BlockSize()])
	out := make([]byte, len(src))
	blockMode.CryptBlocks(out, src)
	strOut := base64.StdEncoding.EncodeToString(out)
	return strOut, nil
}

// DesDecryptFromBase base64 decoding, and decrypt with priKey
func DesDecryptFromBase(crypted, key []byte) ([]byte, error) {
	ori, _ := base64.StdEncoding.DecodeString(string(crypted))
	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(block, key[:block.BlockSize()])
	out := make([]byte, len(ori))
	blockMode.CryptBlocks(out, ori)
	out = PKCS5UnPadding(out)
	return out, nil
}

// PKCS7Padding
func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// PKCS7UnPadding
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
	return base64.StdEncoding.EncodeToString(origData), nil
}
