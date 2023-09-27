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

package tools

import (
	"fmt"
	"testing"
)

// TestAesDeEncrytion aes DeEncrytion test
func TestAesEnDecrytion(t *testing.T) {
	// 需要16的倍数
	priKey, err := randStr(32)
	if err != nil {
		t.Errorf("randStr err: %s\n", err.Error())
	}
	oriStr, err := randStr(32)
	if err != nil {
		t.Errorf("randStr err: %s\n", err.Error())
	}
	fmt.Println("original: ", oriStr)
	encrypted, err := AesEncrypt([]byte(oriStr), []byte(priKey))
	if err != nil {
		t.Errorf("encrypt err: %s\n", err.Error())
	}
	fmt.Println("encryptd: ", encrypted)

	original, err := AesDecrypt(encrypted, []byte(priKey))
	if err != nil {
		t.Errorf("decrypt err: %s\n", err.Error())
	}
	fmt.Println("decryptd: ", original)
	if original != oriStr {
		t.Errorf("Decryption Error, old: %s, new: %s", oriStr, original)
	}
}

func TestEnDecryptCredential(t *testing.T) {
	priKey, err := randStr(32)
	if err != nil {
		t.Errorf("randStr err: %s\n", err.Error())
	}
	oriStr, err := randStr(32)
	if err != nil {
		t.Errorf("randStr err: %s\n", err.Error())
	}
	algo := "aes"
	fmt.Println("original: ", oriStr)
	encrypted, err := EncryptCredential(oriStr, priKey, algo)
	if err != nil {
		t.Errorf("encrypt err: %s\n", err.Error())
		t.Fail()
	}
	fmt.Println("encryptd: ", encrypted)

	decryptd, err := DecryptCredential(encrypted, priKey, algo)
	if err != nil {
		t.Errorf("decrypt err: %s\n", err.Error())
		t.Fail()
	}
	fmt.Println("decryptd: ", decryptd)
	if decryptd != oriStr {
		t.Errorf("Decryption Error, old: %s, new: %s", oriStr, decryptd)
		t.Fail()
	}
}
