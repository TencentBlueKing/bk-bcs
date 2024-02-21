// nolint
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

// Package encryptv2 xxx
package encryptv2

import (
	"encoding/base64"
	"errors"
	"fmt"

	bkcrypto "github.com/TencentBlueKing/crypto-golang-sdk" // nolint
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

	return NewBkCrypto(conf)
}

// cmdbBkCrypto blue king crypto
type bcsBkCrypto struct {
	conf *Config
}

// NewBkCrypto new cmdb crypto from bk crypto
func NewBkCrypto(conf *Config) (Cryptor, error) {
	return &bcsBkCrypto{conf: conf}, nil
}

// Encrypt plaintext
func (c *bcsBkCrypto) Encrypt(plaintext string) (string, error) {
	crypto, err := newBkCrypto(c.conf)
	if err != nil {
		return "", err
	}
	ciphertext, err := crypto.Encrypt([]byte(plaintext))
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
	crypto, err := newBkCrypto(c.conf)
	if err != nil {
		return "", err
	}
	plaintext, err := crypto.Decrypt(cipherBytes)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func newBkCrypto(conf *Config) (bkcrypto.Crypto, error) {
	var (
		bkCrypto bkcrypto.Crypto
		err      error
	)

	switch conf.Algorithm {
	case Sm4:
		bkCrypto, err = NewSm4Crypto(conf.Sm4)
	case AesGcm:
		bkCrypto, err = NewAesGcmCrypto(conf.AesGcm)
	case Normal:
		bkCrypto, err = NewCommonCrypto(conf.Normal)
	default:
		return nil, fmt.Errorf("crypto algorithm %s is invalid", conf.Algorithm)
	}

	if err != nil {
		return nil, fmt.Errorf("init %s crypto failed, err: %v", conf.Algorithm, err)
	}

	return bkCrypto, nil
}

type aesGcmCrypto struct {
	kv       string
	nonce    string
	randomIv string // nolint
}

// NewAesGcmCrypto bcs crypto from aesFcm, kv 16,24,32 && nonce 12-16
func NewAesGcmCrypto(aes *AesGcmConf) (bkcrypto.Crypto, error) {
	return &aesGcmCrypto{
		kv:    aes.Key,
		nonce: aes.Nonce,
	}, nil
}

// Encrypt plaintext
func (c *aesGcmCrypto) Encrypt(plaintext []byte) ([]byte, error) {
	var (
		randVal []byte
	)
	if c.nonce != "" {
		randVal = []byte(c.nonce)
	} else {
		randVal = []byte(RandomId(algorithmRandLen[AesGcm]))
	}

	crypto, err := bkcrypto.NewAesGcm([]byte(c.kv), randVal)
	if err != nil {
		return nil, err
	}

	ciphertext, err := crypto.Encrypt(plaintext)
	if err != nil {
		return nil, err
	}
	if c.nonce != "" {
		return ciphertext, nil
	}

	// append randVal to the front of the ciphertext
	ciphertext = append(randVal, ciphertext...)
	return ciphertext, nil
}

// Decrypt ciphertext
func (c *aesGcmCrypto) Decrypt(ciphertext []byte) ([]byte, error) {
	if c.nonce != "" {
		crypto, err := bkcrypto.NewAesGcm([]byte(c.kv), []byte(c.nonce))
		if err != nil {
			return nil, err
		}
		plaintext, err := crypto.Decrypt(ciphertext)
		if err != nil {
			return nil, err
		}

		return plaintext, err
	}

	// get randVal from the front of the ciphertext
	randValLen := algorithmRandLen[AesGcm]
	randVal := ciphertext[:randValLen]

	crypto, err := bkcrypto.NewAesGcm([]byte(c.kv), randVal)
	if err != nil {
		return nil, err
	}
	plaintext, err := crypto.Decrypt(ciphertext[randValLen:])
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

type sm4Crypto struct {
	kv       string
	iv       string
	randomIv string // nolint
}

// NewSm4Crypto bcs crypto from sm4， kv 16 && iv 16
func NewSm4Crypto(sm4 *Sm4Conf) (bkcrypto.Crypto, error) {
	return &sm4Crypto{
		kv: sm4.Key,
		iv: sm4.Iv,
	}, nil
}

// Encrypt plaintext
func (c *sm4Crypto) Encrypt(plaintext []byte) ([]byte, error) {
	var (
		randVal []byte
	)
	if c.iv != "" {
		randVal = []byte(c.iv)
	} else {
		randVal = []byte(RandomId(algorithmRandLen[Sm4]))
	}

	crypto, err := bkcrypto.NewSm4([]byte(c.kv), randVal)
	if err != nil {
		return nil, err
	}

	ciphertext, err := crypto.Encrypt(plaintext)
	if err != nil {
		return nil, err
	}
	// append randVal to the front of the ciphertext
	if c.iv != "" {
		return ciphertext, nil
	}

	ciphertext = append(randVal, ciphertext...)
	return ciphertext, nil
}

// Decrypt ciphertext
func (c *sm4Crypto) Decrypt(ciphertext []byte) ([]byte, error) {
	if c.iv != "" {
		crypto, err := bkcrypto.NewSm4([]byte(c.kv), []byte(c.iv))
		if err != nil {
			return nil, err
		}
		plaintext, err := crypto.Decrypt(ciphertext)
		if err != nil {
			return nil, err
		}

		return plaintext, err
	}

	// get randVal from the front of the ciphertext
	randValLen := algorithmRandLen[Sm4]
	randVal := ciphertext[:randValLen]

	crypto, err := bkcrypto.NewSm4([]byte(c.kv), randVal)
	if err != nil {
		return nil, err
	}
	plaintext, err := crypto.Decrypt(ciphertext[randValLen:])
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

// commonCrypto bcs default crypto
type commonCrypto struct {
	compileKey string
	priKey     string
}

// NewCommonCrypto bcs crypto from common
func NewCommonCrypto(normal *NormalConf) (bkcrypto.Crypto, error) {
	return &commonCrypto{
		compileKey: normal.CompileKey,
		priKey:     normal.PriKey,
	}, nil
}

// Encrypt plaintext
func (c *commonCrypto) Encrypt(plaintext []byte) ([]byte, error) {
	ciphertext, err := DesEncryptToBaseV2(plaintext, func() string {
		if len(c.compileKey) != 0 {
			return c.compileKey
		}

		return c.priKey
	}())
	if err != nil {
		return nil, err
	}

	return ciphertext, nil
}

// Decrypt ciphertext
func (c *commonCrypto) Decrypt(ciphertext []byte) ([]byte, error) {
	plaintext, err := DesDecryptFromBaseV2(ciphertext, func() string {
		if len(c.compileKey) != 0 {
			return c.compileKey
		}

		return c.priKey
	}())
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
