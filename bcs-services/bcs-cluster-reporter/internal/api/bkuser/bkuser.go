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

// Package bkuser bk-user client
package bkuser

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/rest"
)

const (
	listTenantPath = "/api/v3/open/tenants/"
	systemTenantID = "system"
	defaultTimeout = 30 * time.Second
	maxBodySize    = 10 << 20
)

// tenant id is used as request header, only allow safe characters
var tenantIDRegexp = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,63}$`)

// Tenant bk-user tenant
type Tenant struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type listTenantResp struct {
	Data []Tenant `json:"data"`
}

type authInfo struct {
	BkAppCode   string `json:"bk_app_code"`
	BkAppSecret string `json:"bk_app_secret"`
}

// Client bk-user client
type Client struct {
	server     string
	userAuth   string
	httpClient *http.Client
}

// NewClient create bk-user client
func NewClient(server, appCode, appSecret string) (*Client, error) {
	auth, err := json.Marshal(authInfo{BkAppCode: appCode, BkAppSecret: appSecret})
	if err != nil {
		return nil, err
	}
	return &Client{
		server:     strings.TrimSuffix(server, "/"),
		userAuth:   string(auth),
		httpClient: &http.Client{Timeout: defaultTimeout},
	}, nil
}

// ListTenant list bk-user tenants, tenants with invalid id are skipped
func (c *Client) ListTenant(ctx context.Context) ([]Tenant, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.server+listTenantPath, nil)
	if err != nil {
		return nil, fmt.Errorf("new bk-user ListTenant request failed: %s", err.Error())
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Bkapi-Authorization", c.userAuth)
	req.Header.Set(rest.BkTenantIdHeaderKey, systemTenantID)

	rsp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call bk-user ListTenant failed: %s", err.Error())
	}
	defer rsp.Body.Close()

	if rsp.StatusCode < http.StatusOK || rsp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("call bk-user ListTenant failed with status: %d", rsp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(rsp.Body, maxBodySize))
	if err != nil {
		return nil, fmt.Errorf("read bk-user ListTenant response failed: %s", err.Error())
	}
	result := &listTenantResp{}
	if err = json.Unmarshal(body, result); err != nil {
		return nil, fmt.Errorf("unmarshal bk-user ListTenant response failed: %s", err.Error())
	}

	tenants := make([]Tenant, 0, len(result.Data))
	for _, tenant := range result.Data {
		if !tenantIDRegexp.MatchString(tenant.ID) {
			klog.Warningf("skip bk-user tenant with invalid id %q", tenant.ID)
			continue
		}
		tenants = append(tenants, tenant)
	}
	return tenants, nil
}
