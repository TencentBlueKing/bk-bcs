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
	"errors"
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

func GenerateRSAKeyPairToString(keySize int) (string, string, error) {
	privateKey, publicKey, err := GenerateRSAKeyPair(keySize)
	if err != nil {
		return "", "", err
	}

	privateKeyStr, err := PrivateKeyToString(privateKey)
	if err != nil {
		return "", "", err
	}

	publicKeyStr, err := PublicKeyToString(publicKey)
	if err != nil {
		return "", "", err
	}

	return privateKeyStr, publicKeyStr, nil

}

func PrivateKeyToString(privateKey *rsa.PrivateKey) (string, error) {
	privateKey.Validate()
	// 将私钥转换成PEM格式
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	// 返回PEM格式的私钥字符串
	return string(privateKeyPEM), nil
}

func PublicKeyToString(publicKey *rsa.PublicKey) (string, error) {
	// 将公钥转换成PEM格式
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", err
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	// 返回PEM格式的公钥字符串
	return string(publicKeyPEM), nil
}

func VerifyRSAPublicKey(publicKey string) error {
	//msg := []byte(plain)
	// 解码公钥
	pubBlock, _ := pem.Decode([]byte(publicKey))
	// 读取公钥
	pubKeyValue, err := x509.ParsePKIXPublicKey(pubBlock.Bytes)
	if err != nil {
		return err
	}

	pub := pubKeyValue.(*rsa.PublicKey)
	// 检查公钥是否有效
	if pub.N == nil || pub.E == 0 {
		return errors.New("invalid RSA public key")
	}

	return nil
}

// EncryptRSAWithPublicKey encrypts data using an RSA public key.
// Parameter pemPublicKey is the PEM formatted RSA public key string.
// Parameter data is the data to be encrypted.
// It returns the encrypted data or an error.
func EncryptRSAWithPublicKey(pemPublicKey []byte, data []byte) ([]byte, error) {
	// 将 PEM 格式的公钥解码
	block, _ := pem.Decode(pemPublicKey)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	// 使用 ParsePKIXPublicKey 方法解析公钥
	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	// 将解析后的公钥转换为 *rsa.PublicKey 类型
	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("failed to convert public key to RSA public key")
	}

	// 使用公钥加密数据
	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, rsaPublicKey, data)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt data: %w", err)
	}

	return ciphertext, nil
}
