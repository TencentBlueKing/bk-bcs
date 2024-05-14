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

// Package client is client for controller
package client

import (
	"fmt"
	"net/http"
	"time"

	resty "github.com/go-resty/resty/v2"
)

const (
	timeout = time.Second * 10

	netserviceURL = "/netservicecontroller/v1/allocator"
)

// NetserviceClient client for netservice
type NetserviceClient struct {
	address string
	cli     *resty.Client
}

// New create netservice client
func New(address string) (*NetserviceClient, error) {
	return &NetserviceClient{
		address: address,
		cli:     resty.New().SetTimeout(timeout),
	}, nil
}

// Allocate allocates ip
func (nc *NetserviceClient) Allocate(req *AllocateReq) (*AllocateResp, error) {
	url := fmt.Sprintf("%s%s", nc.address, netserviceURL)
	aresp := &AllocateResp{}
	respRest, restErr := nc.cli.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		SetResult(aresp).
		Post(url)
	if restErr != nil {
		return nil, restErr
	}
	if respRest.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("http code %d != 200", respRest.StatusCode())
	}
	return aresp, nil
}

// Release releases ip
func (nc *NetserviceClient) Release(req *ReleaseReq) (*ReleaseResp, error) {
	url := fmt.Sprintf("%s%s", nc.address, netserviceURL)
	rresp := &ReleaseResp{}
	respRest, restErr := nc.cli.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		SetResult(rresp).
		Delete(url)
	if restErr != nil {
		return nil, restErr
	}
	if respRest.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("http code %d != 200", respRest.StatusCode())
	}
	return rresp, nil
}
