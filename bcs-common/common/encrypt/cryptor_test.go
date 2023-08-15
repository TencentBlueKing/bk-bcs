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
	"testing"
)

func TestNewCrypto(t *testing.T) {
	t.Log("normal")
	normalCryptor(t)

	t.Log("sm4")
	sm4Cryptor(t)
}

func normalCryptor(t *testing.T) {
	cfg := &Config{
		Enabled:   true,
		Algorithm: Normal,
		PriKey:    "",
	}
	cry, _ := NewCrypto(cfg)
	a, _ := cry.Encrypt("xxx")
	t.Log(a)
	b, _ := cry.Decrypt(a)
	t.Log(b)

	a, _ = cry.Encrypt("yyy")
	t.Log(a)
	b, _ = cry.Decrypt(a)
	t.Log(b)
}

func sm4Cryptor(t *testing.T) {
	cfg := &Config{
		Enabled:   true,
		Algorithm: Sm4,
		Sm4: &Sm4Conf{
			Key: "xxx",
			Iv:  "xxx",
		},
	}
	cry, err := NewCrypto(cfg)
	if err != nil {
		t.Fatal(err)
	}

	a, _ := cry.Encrypt("Blueking@xxx")
	t.Log(a)
	b, _ := cry.Decrypt(a)
	t.Log(b)

	a, _ = cry.Encrypt("Blueking@xxx")
	t.Log(a)
	b, _ = cry.Decrypt(a)
	t.Log(b)
}
