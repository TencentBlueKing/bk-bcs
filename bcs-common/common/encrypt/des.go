/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package encrypt

import (
	"bytes"
	"crypto/cipher"
	"crypto/des" // nolint
	"encoding/base64"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
)

var (
	// key for encryption
	priKey = static.EncryptionKey
)

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

// EncryptOrigin 加密方法
func EncryptOrigin(src []byte, priKey string) ([]byte, error) {
	block, err := des.NewTripleDESCipher([]byte(priKey)) // nolint
	if err != nil {
		return nil, err
	}
	src = PKCS5Padding(src, block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, []byte(priKey)[:block.BlockSize()])
	out := make([]byte, len(src))
	blockMode.CryptBlocks(out, src)

	return out, nil
}

// DecryptOrigin 解密方法
func DecryptOrigin(src []byte, priKey string) ([]byte, error) {
	block, err := des.NewTripleDESCipher([]byte(priKey)) // nolint
	if err != nil {
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(block, []byte(priKey)[:block.BlockSize()])
	out := make([]byte, len(src))
	blockMode.CryptBlocks(out, src)
	out = PKCS5UnPadding(out)
	return out, nil
}

// Encrypt 加密方法
func Encrypt(src []byte, priKey string) ([]byte, error) {
	out, err := EncryptOrigin(src, priKey)
	if err != nil {
		return nil, err
	}
	strOut := base64.StdEncoding.EncodeToString(out)
	return []byte(strOut), nil
}

// Decrypt 解密方法
func Decrypt(src []byte, priKey string) ([]byte, error) {
	ori, _ := base64.StdEncoding.DecodeString(string(src))
	return DecryptOrigin(ori, priKey)
}

// DesEncryptToBaseV2 encrypt with priKey simply, out base64 string
func DesEncryptToBaseV2(src []byte, secretKey string) ([]byte, error) {
	if len(priKey) != 0 {
		return EncryptOrigin(src, priKey)
	}

	if len(secretKey) != 0 {
		return EncryptOrigin(src, secretKey)
	}

	return src, nil
}

// DesDecryptFromBaseV2 base64 decoding, and decrypt with priKey.
func DesDecryptFromBaseV2(src []byte, secretKey string) ([]byte, error) {
	if len(priKey) != 0 {
		return DecryptOrigin(src, priKey)
	}

	if len(secretKey) != 0 {
		return DecryptOrigin(src, secretKey)
	}

	return src, nil
}

// DesEncryptToBase encrypt with priKey simply, out base64 string
func DesEncryptToBase(src []byte) ([]byte, error) {
	if len(priKey) != 0 {
		return Encrypt(src, priKey)
	}
	return src, nil
}

// DesDecryptFromBase base64 decoding, and decrypt with priKey
func DesDecryptFromBase(src []byte) ([]byte, error) {
	if len(priKey) != 0 {
		return Decrypt(src, priKey)
	}
	return src, nil
}
