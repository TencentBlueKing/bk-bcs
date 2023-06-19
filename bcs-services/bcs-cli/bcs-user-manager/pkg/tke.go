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
 *
 */

package pkg

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

const (
	addTkeCidrUrl                = "/v1/tke/cidr/add_cidr"
	applyTkeCidrUrl              = "/v1/tke/cidr/apply_cidr"
	releaseTkeCidrUrl            = "/v1/tke/cidr/release_cidr"
	listTkeCidrUrl               = "/v1/tke/cidr/list_count"
	syncTkeClusterCredentialsUrl = "/v1/tke/%s/sync_credentials"
)

// TkeCidr xxx
type TkeCidr struct {
	Cidr     string `json:"cidr" validate:"required"`
	IpNumber uint   `json:"ip_number" validate:"required"`
	Status   string `json:"status"`
}

// AddTkeCidrForm xxx
type AddTkeCidrForm struct {
	Vpc      string    `json:"vpc" validate:"required"`
	TkeCidrs []TkeCidr `json:"tke_cidrs" validate:"required"`
}

// AddTkeCidrResponse is the response of AddTkeCidr
type AddTkeCidrResponse struct {
	Result  bool     `json:"result"`
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    *TkeCidr `json:"data"`
}

// AddTkeCidr init tke cidrs from bcs-user-manager
func (c *UserManagerClient) AddTkeCidr(reqBody string) (*AddTkeCidrResponse, error) {
	reqForm := new(AddTkeCidrForm)
	if err := json.Unmarshal([]byte(reqBody), reqForm); err != nil {
		return nil, errors.Wrapf(err, "tke cidrs from json unmarshal failed with '%s'", reqBody)
	}
	bs, err := c.do(addTkeCidrUrl, http.MethodPost, nil, reqForm)
	if err != nil {
		return nil, errors.Wrapf(err, "add tke cidr with '%s' failed", reqForm)
	}
	resp := new(AddTkeCidrResponse)
	if err := json.Unmarshal(bs, resp); err != nil {
		return nil, errors.Wrapf(err, "add tke cidr unmarshal failed with response '%s'", string(bs))
	}
	return resp, nil
}

// ApplyTkeCidrForm xxx
type ApplyTkeCidrForm struct {
	Vpc      string `json:"vpc" validate:"required"`
	Cluster  string `json:"cluster" validate:"required"`
	IpNumber uint   `json:"ip_number" validate:"required"`
}

// ApplyTkeCidrResult xxx
type ApplyTkeCidrResult struct {
	Vpc      string `json:"vpc" validate:"required"`
	Cidr     string `json:"cidr" validate:"required"`
	IpNumber uint   `json:"ip_number" validate:"required"`
	Status   string `json:"status"`
}

// ApplyTkeCidrResponse is the response of ApplyTkeCidr
type ApplyTkeCidrResponse struct {
	Result  bool                `json:"result"`
	Code    int                 `json:"code"`
	Message string              `json:"message"`
	Data    *ApplyTkeCidrResult `json:"data"`
}

// ApplyTkeCidr  assign a cidr to client from bcs-user-manager
func (c *UserManagerClient) ApplyTkeCidr(reqBody string) (*ApplyTkeCidrResponse, error) {
	reqForm := new(ApplyTkeCidrForm)
	if err := json.Unmarshal([]byte(reqBody), reqForm); err != nil {
		return nil, errors.Wrapf(err, "form json unmarshal failed with '%s'", reqBody)
	}
	bs, err := c.do(applyTkeCidrUrl, http.MethodPost, nil, reqBody)
	if err != nil {
		return nil, errors.Wrapf(err, "apply tke cidr with '%s' failed", reqBody)
	}
	resp := new(ApplyTkeCidrResponse)
	if err := json.Unmarshal(bs, resp); err != nil {
		return nil, errors.Wrapf(err, "apply tke cidr unmarshal failed with response '%s'", string(bs))
	}
	return resp, nil
}

// ReleaseTkeCidrForm xxx
type ReleaseTkeCidrForm struct {
	Vpc     string `json:"vpc" validate:"required"`
	Cidr    string `json:"cidr" validate:"required"`
	Cluster string `json:"cluster" validate:"required"`
}

// ReleaseTkeCidrResponse is the response of ReleaseTkeCidr
type ReleaseTkeCidrResponse struct {
	Result  bool     `json:"result"`
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    *TkeCidr `json:"data"`
}

// ReleaseTkeCidr  release a cidr to client from bcs-user-manager
func (c *UserManagerClient) ReleaseTkeCidr(reqBody string) (*ReleaseTkeCidrResponse, error) {
	reqForm := new(ReleaseTkeCidrForm)
	if err := json.Unmarshal([]byte(reqBody), reqForm); err != nil {
		return nil, errors.Wrapf(err, "release tke cidrs form json unmarshal failed with '%s'", reqBody)
	}
	bs, err := c.do(releaseTkeCidrUrl, http.MethodPost, nil, reqForm)
	if err != nil {
		return nil, errors.Wrapf(err, "release tke cidr with '%s' failed", reqForm)
	}
	resp := new(ReleaseTkeCidrResponse)
	if err := json.Unmarshal(bs, resp); err != nil {
		return nil, errors.Wrapf(err, "release tke cidr unmarshal failed with response '%s'", string(bs))
	}
	return resp, nil
}

// CidrCount cidrInfo
type CidrCount struct {
	Count    int    `json:"count"`
	Vpc      string `json:"vpc"`
	IpNumber uint   `json:"ip_number"`
	Status   string `json:"status"`
}

// ListTkeCidrResponse is the response of ListTkeCidr
type ListTkeCidrResponse struct {
	Result  bool        `json:"result"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    []CidrCount `json:"data"`
}

// ListTkeCidr  list cidr count group by vpc from bcs-user-manager
func (c *UserManagerClient) ListTkeCidr() (*ListTkeCidrResponse, error) {
	bs, err := c.do(listTkeCidrUrl, http.MethodPost, nil, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "list cidr count group by vpc failed")
	}
	resp := new(ListTkeCidrResponse)
	if err := json.Unmarshal(bs, resp); err != nil {
		return nil, errors.Wrapf(err, "list cidr count group by vpc unmarshal failed with response '%s'", string(bs))
	}
	return resp, nil
}

// SyncTkeClusterCredentialsResponse is the response of ReleaseTkeCidr
type SyncTkeClusterCredentialsResponse struct {
	Result  bool     `json:"result"`
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    *TkeCidr `json:"data"`
}

// SyncTkeClusterCredentials  sync the tke cluster credentials from bcs-user-manager
func (c *UserManagerClient) SyncTkeClusterCredentials(clusterID string) (*SyncTkeClusterCredentialsResponse, error) {
	bs, err := c.do(fmt.Sprintf(syncTkeClusterCredentialsUrl, clusterID), http.MethodPost, nil, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "sync the tke cluster credentials with clusterID '%s' failed", clusterID)
	}
	resp := new(SyncTkeClusterCredentialsResponse)
	if err := json.Unmarshal(bs, resp); err != nil {
		return nil, errors.Wrapf(err, "sync the tke cluster credentials unmarshal failed with response '%s'", string(bs))
	}
	return resp, nil
}
