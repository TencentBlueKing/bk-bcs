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

package service

import (
	"context"
	"encoding/base64"
	"testing"

	"bscp.io/pkg/tools"

	"github.com/hashicorp/vault/sdk/logical"
)

func TestKvEncryptRsa(t *testing.T) {

	// 1.上传 kv
	var resp *logical.Response
	var err error
	b, storage := createBackendWithStorage(t)

	kvPath := "apps/1/kvs/1"

	req := &logical.Request{
		Path:      kvPath,
		Operation: logical.CreateOperation,
		Data: map[string]interface{}{
			"value": "MQ==",
		},
		Storage: storage,
	}

	resp, err = b.HandleRequest(context.Background(), req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("err:%v resp:%#v", err, resp)
	}

	// 2.上传公钥
	keyPath := "apps/1/keys/2"

	req = &logical.Request{
		Operation: logical.CreateOperation,
		Path:      keyPath,
		Storage:   storage,
		Data: map[string]interface{}{
			"algorithm": "rsa",
		},
	}
	resp, err = b.HandleRequest(context.Background(), req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("err:%v resp:%#v", err, resp)
	}

	// 解析出私钥
	privateKeyPEM, ok := resp.Data["private_key"].(string)
	if !ok {
		t.Fatalf("failed to get key:%v resp:%#v", err, resp)
	}

	kvEncrypt := "apps/1/kvs/1/encrypt"
	req = &logical.Request{
		Operation: logical.CreateOperation,
		Path:      kvEncrypt,
		Storage:   storage,
		Data: map[string]interface{}{
			"algorithm": "rsa",
			"key_name":  "2",
		},
	}
	resp, err = b.HandleRequest(context.Background(), req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("err:%v resp:%#v", err, resp)
	}

	ciphertextBase64, ok := resp.Data["ciphertext"].(string)
	if !ok {
		t.Fatalf("failed to get ciphertext:%v resp:%#v", err, resp)
	}

	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)

	privateKey, err := tools.RSAPrivateKeyFromPEM([]byte(privateKeyPEM))
	if err != nil {
		t.Fatalf("err:%v resp:%#v", err, resp)
	}

	plaintext, err := tools.RSADecryptWithPrivateKey(privateKey, []byte(ciphertext))
	if err != nil {
		t.Fatalf("err:%v resp:%#v", err, resp)
	}
	if string(plaintext) != "MQ==" {
		t.Fatalf("err:%v resp:%#v", err, resp)
	}

}
