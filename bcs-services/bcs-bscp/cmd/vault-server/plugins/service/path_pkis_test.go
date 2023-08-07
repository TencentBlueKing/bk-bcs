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

// 上传RSA密钥
// 上传一个公钥
// 获取这个公钥
// 更新这个公钥
// 获取公钥
// 删除公钥
func TestPkiRsa(t *testing.T) {

	var resp *logical.Response
	var err error
	b, s := createBackendWithStorage(t)

	key1 := "-----BEGIN RSA PUBLIC KEY-----\nMIGJAoGBAMfmy5Zz8Y7NE+uEHrfOwN1FaT0INVIFWJH5IWCihOAJRElMX6zVAr2j\nKxSaiQVQzf5SDYj1xDLlaBk2D+Ygl2DAR5C7sCULjSX8Ok4TnVMS/puZQLctBTSP\nixSoDGTejm7JUU2zZdBBRWw2q7+D+UoERqEXLCTTyFT8a1wcM/11AgMBAAE=\n-----END RSA PUBLIC KEY-----"

	key2 := "-----BEGIN RSA PUBLIC KEY-----\nMIGJAoGBALT0avOadBoMLYEdwR6fHGVcKo7zZ7f+y77f1KA9xxsDG4LOxLlEqgK3\nC1LdAjguz7C0TA5ayEhv9uojs6nfVQc/5dw2ZVoqpYiNguZBFcsMHb+Oqgi9qhHq\n7ZWH5KyzfVmZkDYWTARQeHp0tNah8I/6Ha1DQ0peTHvX9YbDsuipAgMBAAE=\n-----END RSA PUBLIC KEY-----\n"

	path := "apps/1/pkis/2"

	req := &logical.Request{
		Operation: logical.CreateOperation,
		Path:      path,
		Storage:   s,
		Data: map[string]interface{}{
			"algorithm": "rsa",
			"pub_key":   key1,
		},
	}
	resp, err = b.HandleRequest(context.Background(), req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("err:%v resp:%#v", err, resp)
	}

	req = &logical.Request{
		Operation: logical.ReadOperation,
		Path:      path,
		Storage:   s,
		Data: map[string]interface{}{
			"algorithm": "rsa",
		},
	}
	resp, err = b.HandleRequest(context.Background(), req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("err:%v resp:%#v", err, resp)
	}
	if resp.Data["pub_key"] != key1 {
		t.Fatalf("获取公钥的错误")
	}

	req = &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      path,
		Storage:   s,
		Data: map[string]interface{}{
			"algorithm": "rsa",
			"pub_key":   key2,
		},
	}
	resp, err = b.HandleRequest(context.Background(), req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("err:%v resp:%#v", err, resp)
	}

	req = &logical.Request{
		Operation: logical.ReadOperation,
		Path:      path,
		Storage:   s,
		Data: map[string]interface{}{
			"algorithm": "rsa",
		},
	}
	resp, err = b.HandleRequest(context.Background(), req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("err:%v resp:%#v", err, resp)
	}
	if resp.Data["pub_key"] != key2 {
		t.Fatalf("获取公钥的错误")
	}

	req = &logical.Request{
		Operation: logical.DeleteOperation,
		Path:      path,
		Storage:   s,
		Data: map[string]interface{}{
			"algorithm": "rsa",
		},
	}
	resp, err = b.HandleRequest(context.Background(), req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("err:%v resp:%#v", err, resp)
	}

}

// 生成SM2密钥
func TestPkiSM2(t *testing.T) {

	var resp *logical.Response
	var err error
	b, s := createBackendWithStorage(t)

	key1 := "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoEcz1UBgi0DQgAEShifZVIvzOIwaFgPEwsd174D5np6\n7B78u37va+rEMfxZJHPvvGw5zyqVA74+zXzwSD2BtVwTf5LQxIB42io2pQ==\n-----END PUBLIC KEY-----"
	key2 := "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoEcz1UBgi0DQgAEATwWX6WEpQIFxipM1s8Qp8F/yAGA\nbPP/1dphzANQOtMkK54L2bkubXaxuCwIN9xBp3vZxNj7sT16yS+swjOunQ==\n-----END PUBLIC KEY-----\n"

	path := "apps/1/pkis/2"

	req := &logical.Request{
		Operation: logical.CreateOperation,
		Path:      path,
		Storage:   s,
		Data: map[string]interface{}{
			"algorithm": "sm2",
			"pub_key":   key1,
		},
	}
	resp, err = b.HandleRequest(context.Background(), req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("err:%v resp:%#v", err, resp)
	}

	req = &logical.Request{
		Operation: logical.ReadOperation,
		Path:      path,
		Storage:   s,
		Data: map[string]interface{}{
			"algorithm": "sm2",
		},
	}
	resp, err = b.HandleRequest(context.Background(), req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("err:%v resp:%#v", err, resp)
	}
	if resp.Data["pub_key"] != key1 {
		t.Fatalf("获取公钥的错误")
	}

	req = &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      path,
		Storage:   s,
		Data: map[string]interface{}{
			"algorithm": "sm2",
			"pub_key":   key2,
		},
	}
	resp, err = b.HandleRequest(context.Background(), req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("err:%v resp:%#v", err, resp)
	}

	req = &logical.Request{
		Operation: logical.ReadOperation,
		Path:      path,
		Storage:   s,
		Data: map[string]interface{}{
			"algorithm": "sm2",
		},
	}
	resp, err = b.HandleRequest(context.Background(), req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("err:%v resp:%#v", err, resp)
	}
	if resp.Data["pub_key"] != key2 {
		t.Fatalf("获取公钥的错误")
	}

	req = &logical.Request{
		Operation: logical.DeleteOperation,
		Path:      path,
		Storage:   s,
		Data: map[string]interface{}{
			"algorithm": "sm2",
		},
	}
	resp, err = b.HandleRequest(context.Background(), req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("err:%v resp:%#v", err, resp)
	}

}
