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
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

// GenerateRSAKeyPair 生成RSA密钥对
// keySize 表示密钥的位数，常用的有 2048 和 4096
// 返回生成的RSA私钥和公钥
func GenerateRSAKeyPair(keySize int) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, nil, err
	}
	return privateKey, &privateKey.PublicKey, nil
}

// RSAPrivateKeyToPEM 将私钥编码为PEM格式
func RSAPrivateKeyToPEM(privateKey *rsa.PrivateKey) []byte {
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	return pem.EncodeToMemory(privateKeyPEM)
}

// RSAPublicKeyToPEM 将公钥编码为PEM格式
func RSAPublicKeyToPEM(publicKey *rsa.PublicKey) []byte {
	publicKeyPEM := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(publicKey),
	}

	return pem.EncodeToMemory(publicKeyPEM)
}

// RSAPrivateKeyFromPEM 解码PEM格式的数据
func RSAPrivateKeyFromPEM(pemData []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	// 解析DER格式的私钥数据
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

// RSAPublicKeyFromPEM 解码PEM格式的数据
func RSAPublicKeyFromPEM(pemData []byte) (*rsa.PublicKey, error) {

	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	// 解析DER格式的公钥数据
	publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return publicKey, nil
}

// RSAEncryptWithPublicKey 使用公钥加密数据
func RSAEncryptWithPublicKey(pubKey *rsa.PublicKey, plaintext []byte) ([]byte, error) {
	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey, plaintext)
	if err != nil {
		return nil, err
	}

	return ciphertext, nil
}

// RSADecryptWithPrivateKey 使用私钥解密数据
func RSADecryptWithPrivateKey(privKey *rsa.PrivateKey, ciphertext []byte) ([]byte, error) {

	plaintext, err := rsa.DecryptPKCS1v15(rand.Reader, privKey, ciphertext)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
