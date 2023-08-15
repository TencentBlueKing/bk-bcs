/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package encrypt

import (
	"encoding/base64"
	"errors"
	"fmt"

	bkcrypto "github.com/TencentBlueKing/crypto-golang-sdk"
)

// Cryptor 密码器
type Cryptor interface {
	// Encrypt 加密方法
	Encrypt(plaintext string) (string, error)
	// Decrypt 解密方法
	Decrypt(ciphertext string) (string, error)
}

// NewCrypto new crypto by config
func NewCrypto(conf *Config) (Cryptor, error) {
	if conf == nil {
		return nil, errors.New("crypto config is nil")
	}

	err := conf.Validate()
	if err != nil {
		return nil, fmt.Errorf("validate crypto config failed, err: %v", err)
	}

	var bkCrypto bkcrypto.Crypto
	switch conf.Algorithm {
	case Sm4:
		bkCrypto, err = bkcrypto.NewSm4([]byte(conf.Sm4.Key), []byte(conf.Sm4.Iv))
	case AesGcm:
		bkCrypto, err = bkcrypto.NewAesGcm([]byte(conf.AesGcm.Key), []byte(conf.AesGcm.Nonce))
	case Normal:
		bkCrypto, err = NewCommonCrypto(conf.PriKey)
	default:
		return nil, fmt.Errorf("crypto algorithm %s is invalid", conf.Algorithm)
	}

	if err != nil {
		return nil, fmt.Errorf("init %s crypto failed, err: %v", conf.Algorithm, err)
	}

	return NewBkCrypto(bkCrypto)
}

// cmdbBkCrypto blue king crypto
type bcsBkCrypto struct {
	crypto bkcrypto.Crypto
}

// NewBkCrypto new cmdb crypto from bk crypto
func NewBkCrypto(crypto bkcrypto.Crypto) (Cryptor, error) {
	if crypto == nil {
		return nil, errors.New("bk crypto is nil")
	}

	return &bcsBkCrypto{
		crypto: crypto,
	}, nil
}

// Encrypt plaintext
func (c *bcsBkCrypto) Encrypt(plaintext string) (string, error) {
	ciphertext, err := c.crypto.Encrypt([]byte(plaintext))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt ciphertext
func (c *bcsBkCrypto) Decrypt(ciphertext string) (string, error) {
	cipherBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	plaintext, err := c.crypto.Decrypt(cipherBytes)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// commonCrypto bcs default crypto
type commonCrypto struct {
	priKey string
}

// NewCommonCrypto bcs crypto from common
func NewCommonCrypto(key string) (bkcrypto.Crypto, error) {
	return &commonCrypto{
		priKey: key,
	}, nil
}

// Encrypt plaintext
func (c *commonCrypto) Encrypt(plaintext []byte) ([]byte, error) {
	ciphertext, err := DesEncryptToBaseV2(plaintext, c.priKey)
	if err != nil {
		return nil, err
	}

	return ciphertext, nil
}

// Decrypt ciphertext
func (c *commonCrypto) Decrypt(ciphertext []byte) ([]byte, error) {
	plaintext, err := DesDecryptFromBaseV2(ciphertext, c.priKey)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
