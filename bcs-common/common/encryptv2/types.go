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

package encryptv2

import (
	"errors"
	"fmt"
	"math/rand"
)

// Config defines cmdb crypto configuration
type Config struct {
	Enabled   bool        `json:"enabled"`
	Algorithm Algorithm   `json:"algorithm"`
	Sm4       *Sm4Conf    `json:"sm4"`
	AesGcm    *AesGcmConf `json:"aes_gcm"`
	Normal    *NormalConf `json:"normal"`
}

// Validate Config
func (conf Config) Validate() error {
	if !conf.Enabled {
		return nil
	}

	switch conf.Algorithm {
	case Sm4:
		if conf.Sm4 == nil {
			return errors.New("sm4 config is not set")
		}
	case AesGcm:
		if conf.AesGcm == nil {
			return errors.New("aes-gcm config is not set")
		}
	case Normal:
		// allow priKey & compileKey is empty && not encrypt && only base64 encoding
	default:
		return fmt.Errorf("crypto algorithm %s is invalid", conf.Algorithm)
	}

	return nil
}

var algorithmRandLen = map[Algorithm]int{
	Sm4:    16,
	AesGcm: 12,
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// RandomId xxx
func RandomId(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))] // nolint
	}
	return string(b)
}

// Algorithm defines cryptography algorithm types
type Algorithm string

const (
	// Sm4 SM4 cryptography algorithm
	Sm4 Algorithm = "SM4"
	// AesGcm AES-GCM cryptography algorithm
	AesGcm Algorithm = "AES-GCM"
	// Normal normal algorithm
	Normal Algorithm = "normal"
)

// String toString
func (a Algorithm) String() string {
	return string(a)
}

// Sm4Conf defines SM4 cryptography algorithm configuration
type Sm4Conf struct {
	Key string `json:"key"`
	Iv  string `json:"iv"`
}

// AesGcmConf defines AES-GCM cryptography algorithm configuration
type AesGcmConf struct {
	Key   string `json:"key"`
	Nonce string `json:"nonce"`
}

// NormalConf define normal algorithm configuration
type NormalConf struct {
	CompileKey string `json:"compileKey"`
	PriKey     string `json:"priKey"`
}
