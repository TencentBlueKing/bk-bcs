/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

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
	// 加密算法
	AES = "aes"
	// 加密长度
	EncryptionLen = 32
)

// CreateCredential create credential
func CreateCredential(masterKey, encryptionAlgorithm string) (string, error) {
	return EncryptCredential(randStr(EncryptionLen), masterKey, encryptionAlgorithm)
}

// EncryptCredential encrypt credential
func EncryptCredential(credential, masterKey, encryptionAlgorithm string) (string, error) {
	if credential == "" {
		return "", fmt.Errorf("credential is null")
	}
	if len(masterKey) == 0 || len(encryptionAlgorithm) == 0 {
		return "", fmt.Errorf("key or encryption algorithm is null")
	}
	switch encryptionAlgorithm {
	case AES:
		return AesEncrypt([]byte(credential), []byte(masterKey))
	default:
		return "", fmt.Errorf("algorithm type is is not supported, type: %s", encryptionAlgorithm)
	}
}

// DecryptCredential Decrypt credential
func DecryptCredential(credential, masterKey, encryptionAlgorithm string) (string, error) {
	if credential == "" {
		return "", fmt.Errorf("credential is null")
	}
	if len(masterKey) == 0 || len(encryptionAlgorithm) == 0 {
		return "", fmt.Errorf("key or encryption algorithm is null")
	}
	switch encryptionAlgorithm {
	case AES:
		return AesDecrypt(credential, []byte(masterKey))
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
func AesDecrypt(crypted string, key []byte) (string, error) {
	b64Byte, err := base64.StdEncoding.DecodeString(crypted)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(b64Byte))
	blockMode.CryptBlocks(origData, b64Byte)
	origData = PKCS7UnPadding(origData)
	return string(origData), nil
}
