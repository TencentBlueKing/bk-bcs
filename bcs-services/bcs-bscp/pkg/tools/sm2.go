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
	"io"

	"github.com/tjfoc/gmsm/sm2"
	"github.com/tjfoc/gmsm/x509"
)

// GenerateSM2KeyPair 生成SM2 密钥对
func GenerateSM2KeyPair(random io.Reader) (*sm2.PrivateKey, error) {
	privateKey, err := sm2.GenerateKey(random)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

// SM2PrivateKeyToPEM 将私钥编码为PEM格式
func SM2PrivateKeyToPEM(privateKey *sm2.PrivateKey) ([]byte, error) {
	privateKeyToPem, err := x509.WritePrivateKeyToPem(privateKey, nil)
	if err != nil {
		return nil, err
	}

	return privateKeyToPem, nil
}

// SM2PublicKeyToPEM 将公钥编码为PEM格式
func SM2PublicKeyToPEM(publicKey *sm2.PublicKey) ([]byte, error) {
	publicKeyPEM, err := x509.WritePublicKeyToPem(publicKey)
	if err != nil {
		return nil, err
	}
	return publicKeyPEM, nil
}

// SM2PrivateKeyFromPEM 解码PEM格式的数据
func SM2PrivateKeyFromPEM(pemData []byte) (*sm2.PrivateKey, error) {

	privateKeyFromPem, err := x509.ReadPrivateKeyFromPem(pemData, nil)
	if err != nil {
		return nil, err
	}

	return privateKeyFromPem, nil
}

// SM2PublicKeyFromPEM 解码PEM格式的数据
func SM2PublicKeyFromPEM(pemData []byte) (*sm2.PublicKey, error) {
	publicKeyFromPem, err := x509.ReadPublicKeyFromPem(pemData)
	if err != nil {
		return nil, err
	}

	return publicKeyFromPem, nil
}

// SM2EncryptWithPublicKey 使用公钥加密数据
func SM2EncryptWithPublicKey(pubKey *sm2.PublicKey, plaintext []byte) ([]byte, error) {
	cipherText, err := pubKey.EncryptAsn1(plaintext, rand.Reader)
	if err != nil {
		return nil, err
	}
	return cipherText, nil
}

// SM2DecryptWithPrivateKey 使用私钥解密数据
func SM2DecryptWithPrivateKey(privateKey *sm2.PrivateKey, ciphertext []byte) ([]byte, error) {
	plaintext, err := privateKey.DecryptAsn1(ciphertext)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
