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

// Package encrypt xxx
package encrypt

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/encryptv2" // nolint
	"github.com/Tencent/bk-bcs/bcs-common/common/static"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
)

// encryptClient global encrypt client
var encryptClient encryptv2.Cryptor

// SetEncryptClient set cmdb client
func SetEncryptClient(config encryptv2.Config) error {
	// default inject makefile encrypt key
	config.Normal.CompileKey = static.EncryptionKey

	cli, err := encryptv2.NewCrypto(&config)
	if err != nil {
		return err
	}

	encryptClient = cli
	return nil
}

// GetEncryptClient get encrypt client
func GetEncryptClient() encryptv2.Cryptor {
	return encryptClient
}

// InitEncryptClient init encrypt client
func InitEncryptClient(encrypt *proto.OriginEncrypt) (encryptv2.Cryptor, error) {
	cfg := &encryptv2.Config{
		Enabled:   true,
		Algorithm: encryptv2.Algorithm(encrypt.EncryptType),
		Sm4: &encryptv2.Sm4Conf{
			Key: encrypt.Kv,
			Iv:  encrypt.Iv,
		},
		AesGcm: &encryptv2.AesGcmConf{
			Key:   encrypt.Kv,
			Nonce: encrypt.Iv,
		},
		Normal: &encryptv2.NormalConf{
			CompileKey: encrypt.Kv,
			PriKey:     encrypt.Iv,
		},
	}
	return encryptv2.NewCrypto(cfg)
}

// Encrypt 加密方法
func Encrypt(encrypt encryptv2.Cryptor, plaintext string) (string, error) {
	if !options.GetGlobalCMOptions().Encrypt.Enabled {
		return plaintext, nil
	}

	if encrypt != nil {
		return encrypt.Encrypt(plaintext)
	}

	return encryptClient.Encrypt(plaintext)
}

// Decrypt 解密方法
func Decrypt(decrypt encryptv2.Cryptor, ciphertext string) (string, error) {
	if !options.GetGlobalCMOptions().Encrypt.Enabled {
		return ciphertext, nil
	}

	if decrypt != nil {
		return decrypt.Decrypt(ciphertext)
	}

	return encryptClient.Decrypt(ciphertext)
}
