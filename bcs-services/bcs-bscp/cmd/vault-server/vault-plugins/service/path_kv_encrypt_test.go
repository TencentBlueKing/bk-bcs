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

package service

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

// 创建kv
// 生成rsa密钥对
// 生成sm密钥对
// 获取rsa加密密文
// 获取sm2加密密文
func TestKvEncryptRsa(t *testing.T) {

	// 1.上传 kv
	var resp *logical.Response
	var err error
	b, storage := createBackendWithStorage(t)

	kvPath := "apps/1/kvs/1" //nolint:goconst
	pkiPath := "apps/1/pkis/1"
	kvEncrypt := "apps/1/kvs/1/encrypt"

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

	// 2.生成rsa
	req = &logical.Request{
		Operation: logical.CreateOperation,
		Path:      pkiPath,
		Storage:   storage,
		Data: map[string]interface{}{
			"algorithm": "rsa",
		},
	}
	resp, err = b.HandleRequest(context.Background(), req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("err:%v resp:%#v", err, resp)
	}
	privateKeyRsaPEM, ok := resp.Data["private_key"].(string)
	if !ok {
		t.Fatalf("获取rsa私钥失败")
	}
	privateKeyRsa, err := tools.RSAPrivateKeyFromPEM([]byte(privateKeyRsaPEM))
	if err != nil {
		t.Fatalf("解析rsa私钥失败")
	}

	// rsa 加密
	req = &logical.Request{
		Operation: logical.CreateOperation,
		Path:      kvEncrypt,
		Storage:   storage,
		Data: map[string]interface{}{
			"algorithm": "rsa",
			"key_name":  "1",
		},
	}
	resp, err = b.HandleRequest(context.Background(), req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("err:%v resp:%#v", err, resp)
	}
	ciphertextBase64, ok := resp.Data["ciphertext"].(string)
	if !ok {
		t.Fatalf("获取rsa密文失败")
	}
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		t.Fatalf("err:%v ", err)
	}
	value, err := tools.RSADecryptWithPrivateKey(privateKeyRsa, ciphertext)
	if err != nil {
		t.Fatalf("err : %v", err)
	}
	if string(value) != "MQ==" {
		t.Fatalf("解密rsa失败")
	}

}

func TestKvEncryptSm2(t *testing.T) {

	// 1.上传 kv
	var resp *logical.Response
	var err error
	b, storage := createBackendWithStorage(t)

	kvPath := "apps/1/kvs/1"
	pkiPath := "apps/1/pkis/1"
	kvEncrypt := "apps/1/kvs/1/encrypt"

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

	// 2.生成sm2
	req = &logical.Request{
		Operation: logical.CreateOperation,
		Path:      pkiPath,
		Storage:   storage,
		Data: map[string]interface{}{
			"algorithm": "sm2",
		},
	}
	resp, err = b.HandleRequest(context.Background(), req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("err:%v resp:%#v", err, resp)
	}
	privateKeySm2PEM, ok := resp.Data["private_key"].(string)
	if !ok {
		t.Fatalf("获取sm2私钥失败")
	}
	privateKeySm2, err := tools.SM2PrivateKeyFromPEM([]byte(privateKeySm2PEM))
	if err != nil {
		t.Fatalf("解析sm2私钥失败")
	}

	// sm2 加密
	req = &logical.Request{
		Operation: logical.CreateOperation,
		Path:      kvEncrypt,
		Storage:   storage,
		Data: map[string]interface{}{
			"algorithm": "sm2",
			"key_name":  "1",
		},
	}
	resp, err = b.HandleRequest(context.Background(), req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("err:%v resp:%#v", err, resp)
	}
	ciphertextBase64, ok := resp.Data["ciphertext"].(string)
	if !ok {
		t.Fatalf("获取sm2密文失败")
	}
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		t.Fatalf("err:%v ", err)
	}
	value, err := tools.SM2DecryptWithPrivateKey(privateKeySm2, ciphertext)
	if err != nil {
		t.Fatalf("err : %v", err)
	}
	if string(value) != "MQ==" {
		t.Fatalf("解密sm2失败")
	}

}
