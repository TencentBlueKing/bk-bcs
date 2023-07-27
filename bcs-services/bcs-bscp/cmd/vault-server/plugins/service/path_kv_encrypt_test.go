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
	"testing"

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
			"type": "RSA",
			//"public_key": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAxnuITzBfSs+5yDEhOTt5\n9kZtDQB0DLsyaKRp3NqBp9f8Uu0uQVSuW5yQRSu7Ned6qiiMvpNFODSAKoBk6LgH\noZbU2xJQlRAAj75npjHJtda65ANURjjuX165zRRrirpZg5KFvJ5m5nx+XKxme514\nv8Rf2dhL0dIjzK45Ew4+DDQhbZ84KywAMkHhL+jN00zJsDQ2npkV7/n2bVx/1mLa\n/aL0fjpUqQ6WwaRshIamD+zYx11+G5NF+E1yInx5bQOOGAKbm+UILpltYLjZi7gR\nEwnJkL3K9S4WUmj0oD7Ivczk8qZwGuAQFovGFK5DG1OuQ0j/BXHCzK+7C3+l+pB7\nuwIDAQAB\n-----END PUBLIC KEY-----",
		},
	}
	resp, err = b.HandleRequest(context.Background(), req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("err:%v resp:%#v", err, resp)
	}

	kvEncrypt := "apps/1/kvs/1/encrypt"
	req = &logical.Request{
		Operation: logical.CreateOperation,
		Path:      kvEncrypt,
		Storage:   storage,
		Data: map[string]interface{}{
			"algorithm": "RSA",
			"key_name":  "2",
		},
	}
	resp, err = b.HandleRequest(context.Background(), req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("err:%v resp:%#v", err, resp)
	}

}
