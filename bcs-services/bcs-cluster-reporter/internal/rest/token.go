/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package rest xxx
package rest

import (
	"crypto/tls"
	"net/http"
)

// BkTenantIdHeaderKey 蓝鲸租户 ID 请求头
const BkTenantIdHeaderKey = "X-Bk-Tenant-Id"

// BcsTransport client
type BcsTransport struct {
	Token string
	// TenantID 非空时通过 X-Bk-Tenant-Id 指定请求的租户
	TenantID string
}

// RoundTrip xxx
func (t *BcsTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Set("accept", "application/json")
	header.Set("Content-Type", "application/json")
	header.Set("Authorization", "Bearer "+t.Token)
	if t.TenantID != "" {
		header.Set(BkTenantIdHeaderKey, t.TenantID)
	}
	req.Header = header

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 设置为 true 来禁用证书验证
		},
	}

	return tr.RoundTrip(req)
}

// TenantTransport 为请求添加 X-Bk-Tenant-Id 请求头
type TenantTransport struct {
	TenantID string
	Base     http.RoundTripper
}

// RoundTrip xxx
func (t *TenantTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set(BkTenantIdHeaderKey, t.TenantID)
	return t.Base.RoundTrip(req)
}
