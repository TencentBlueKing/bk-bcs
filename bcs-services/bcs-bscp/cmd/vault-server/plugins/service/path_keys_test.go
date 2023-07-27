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
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
)

func IsValidRSAPublicKey(publicKeyStr string) bool {
	// 将字符串格式的公钥解码为PEM数据
	block, _ := pem.Decode([]byte(publicKeyStr))
	if block == nil || block.Type != "RSA PUBLIC KEY" {
		// 解码失败或者类型不匹配，公钥不合法
		return false
	}

	// 解析PEM数据中的公钥
	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		// 解析失败，公钥不合法
		return false
	}

	// 判断公钥是否为RSA公钥
	_, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		// 不是RSA公钥，公钥不合法
		return false
	}

	// 公钥合法
	return true
}

// 上传RSA密钥
// 上传一个公钥
// 获取这个公钥
// 更新这个公钥
// 获取公钥
// 删除公钥
func Test_case1(t *testing.T) {

	var resp *logical.Response
	var err error
	b, s := createBackendWithStorage(t)

	path := "apps/1/keys/2"

	req := &logical.Request{
		Operation: logical.CreateOperation,
		Path:      path,
		Storage:   s,
		Data: map[string]interface{}{
			"algorithm":  "RSA",
			"public_key": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAxnuITzBfSs+5yDEhOTt5\n9kZtDQB0DLsyaKRp3NqBp9f8Uu0uQVSuW5yQRSu7Ned6qiiMvpNFODSAKoBk6LgH\noZbU2xJQlRAAj75npjHJtda65ANURjjuX165zRRrirpZg5KFvJ5m5nx+XKxme514\nv8Rf2dhL0dIjzK45Ew4+DDQhbZ84KywAMkHhL+jN00zJsDQ2npkV7/n2bVx/1mLa\n/aL0fjpUqQ6WwaRshIamD+zYx11+G5NF+E1yInx5bQOOGAKbm+UILpltYLjZi7gR\nEwnJkL3K9S4WUmj0oD7Ivczk8qZwGuAQFovGFK5DG1OuQ0j/BXHCzK+7C3+l+pB7\nuwIDAQAB\n-----END PUBLIC KEY-----",
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
	}
	resp, err = b.HandleRequest(context.Background(), req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("err:%v resp:%#v", err, resp)
	}

	req = &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      path,
		Storage:   s,
		Data: map[string]interface{}{
			"type":       "RSA",
			"public_key": "-----BEGIN PUBLIC KEY-----\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAmEF5sO5QjJFKbQwGALAp\n8dDwzkc4poilF2NmzxxmUFL8PYYArH72YerZwSTV9frE3RoeusVp1qHji4oGkTbF\nYSKLnYApHGhluJTY4/IqrPHtj4vdaGlpbd3W7Ww1IJU8DOkarP37d+EwXhPGk0Y9\nJmVrccs1OFlX/cY9j15ZHXsXubehnSNHYO/CWg7T3AkquWFJupaeDRyv2mjuvfOM\nMxVW7DR/VD3iYwOYPC7GstX6st/u1vASAgrFlyu6mpv4tvdOxZFaD76DUpzKHQLx\nhy7GVWgLCvSbRu6N6LE8+YnypWOIwtt62a9VcpPsKdfvSbDbLYtRQ2PfLLRlLZrb\nUZEQjv5mSagM3wGNsbYSGVK5ppd4Yyr/Y54QvInCdyUJ3zt5M0VsytsRg+anv1gv\nfEZ0l3a7B0dJcZwTmZiXjcOWg9Zly497sgfUFM6EQwFBRx6dWBCZeqlf0fmAxZ2D\ncCYPQZ9JwugvfrIsI/1fG29s7zJ9XNtmj6fOFBVJTM+7wRkgTBbbxqIlzaP6hBic\n/iTj+Wj4HWQpRlFOFl34LSORFWWP0FMTz+oQo+cjUfcdVy9u9SKxWCXdNynP2fVX\nRTYEQApfUQdw4iMF+qnqDfInTXE4lmef35AyYco4bYc4yR3VPMuQivbXPiGyEVAr\nTMbd5EvZgiqVwpFdW5kBFCsCAwEAAQ==\n-----END PUBLIC KEY-----\n",
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
	}
	resp, err = b.HandleRequest(context.Background(), req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("err:%v resp:%#v", err, resp)
	}

	req = &logical.Request{
		Operation: logical.DeleteOperation,
		Path:      path,
		Storage:   s,
	}
	resp, err = b.HandleRequest(context.Background(), req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("err:%v resp:%#v", err, resp)
	}

}

// 生成RSA密钥
func Test_case2(t *testing.T) {
	var resp *logical.Response
	var err error
	b, s := createBackendWithStorage(t)

	path := "apps/1/keys/2"

	req := &logical.Request{
		Operation: logical.CreateOperation,
		Path:      path,
		Storage:   s,
		Data: map[string]interface{}{
			"type":       "RSA",
			"public_key": "",
			"length":     "2048",
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
	}
	resp, err = b.HandleRequest(context.Background(), req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("err:%v resp:%#v", err, resp)
	}

	req = &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      path,
		Storage:   s,
		Data: map[string]interface{}{
			"type":       "RSA",
			"public_key": "-----BEGIN PUBLIC KEY-----\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAmEF5sO5QjJFKbQwGALAp\n8dDwzkc4poilF2NmzxxmUFL8PYYArH72YerZwSTV9frE3RoeusVp1qHji4oGkTbF\nYSKLnYApHGhluJTY4/IqrPHtj4vdaGlpbd3W7Ww1IJU8DOkarP37d+EwXhPGk0Y9\nJmVrccs1OFlX/cY9j15ZHXsXubehnSNHYO/CWg7T3AkquWFJupaeDRyv2mjuvfOM\nMxVW7DR/VD3iYwOYPC7GstX6st/u1vASAgrFlyu6mpv4tvdOxZFaD76DUpzKHQLx\nhy7GVWgLCvSbRu6N6LE8+YnypWOIwtt62a9VcpPsKdfvSbDbLYtRQ2PfLLRlLZrb\nUZEQjv5mSagM3wGNsbYSGVK5ppd4Yyr/Y54QvInCdyUJ3zt5M0VsytsRg+anv1gv\nfEZ0l3a7B0dJcZwTmZiXjcOWg9Zly497sgfUFM6EQwFBRx6dWBCZeqlf0fmAxZ2D\ncCYPQZ9JwugvfrIsI/1fG29s7zJ9XNtmj6fOFBVJTM+7wRkgTBbbxqIlzaP6hBic\n/iTj+Wj4HWQpRlFOFl34LSORFWWP0FMTz+oQo+cjUfcdVy9u9SKxWCXdNynP2fVX\nRTYEQApfUQdw4iMF+qnqDfInTXE4lmef35AyYco4bYc4yR3VPMuQivbXPiGyEVAr\nTMbd5EvZgiqVwpFdW5kBFCsCAwEAAQ==\n-----END PUBLIC KEY-----\n",
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
	}
	resp, err = b.HandleRequest(context.Background(), req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("err:%v resp:%#v", err, resp)
	}

	req = &logical.Request{
		Operation: logical.DeleteOperation,
		Path:      path,
		Storage:   s,
	}
	resp, err = b.HandleRequest(context.Background(), req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("err:%v resp:%#v", err, resp)
	}
}

// 上传
