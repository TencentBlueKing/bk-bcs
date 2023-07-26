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

func GenerateSM2KeyPair(random io.Reader) (*sm2.PrivateKey, error) {
	privateKey, err := sm2.GenerateKey(random)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

func GenerateSM2KeyPairToString(random io.Reader) (string, string, error) {

	privateKey, err := GenerateSM2KeyPair(random)
	if err != nil {
		return "", "", err
	}

	privateKeyToPem, err := x509.WritePrivateKeyToPem(privateKey, nil)
	if err != nil {
		return "", "", err
	}

	//4.进行SM2公钥断言
	publicKey := privateKey.Public().(*sm2.PublicKey)
	//5.将公钥通过x509序列化并进行pem编码

	publicKeyToPem, err := x509.WritePublicKeyToPem(publicKey)
	if err != nil {
		return "", "", err
	}

	return string(privateKeyToPem), string(publicKeyToPem), nil

}

// 加密
func EncryptSM2(plainText []byte, publicKey []byte) ([]byte, error) {

	publicKeyFromPem, err := x509.ReadPublicKeyFromPem(publicKey)
	if err != nil {
		return nil, err
	}
	cipherByte, err := publicKeyFromPem.EncryptAsn1(plainText, rand.Reader)
	if err != nil {
		return nil, err
	}
	return cipherByte, nil
}
