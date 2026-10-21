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
 */

// Package common xxx
package common

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
)

// rsaCryptoLabel 与节点管理服务端约定的 OAEP 标签
var rsaCryptoLabel = []byte("com.example.crypto.rsa.v1")

// encryptCreditWithPublicKey 将明文凭据用节点管理服务器 RSA 公钥加密，
// 与标准运维/蓝鲸节点管理 V3 的加密方式保持一致：
//  1. RSA-OAEP(SHA256, MGF1(SHA256))，label = rsaCryptoLabel
//  2. 密文前置版本字节 0x01
//  3. 整体 base64 编码
func encryptCreditWithPublicKey(publicKeyPEM, plainText string) (string, error) {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return "", fmt.Errorf("failed to decode public key pem")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse public key: %v", err)
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("public key is not rsa public key")
	}

	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, rsaPub, []byte(plainText), rsaCryptoLabel)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt credit with rsa: %v", err)
	}

	// prepend version byte 0x01
	out := append([]byte{0x01}, ciphertext...)
	return base64.StdEncoding.EncodeToString(out), nil
}
