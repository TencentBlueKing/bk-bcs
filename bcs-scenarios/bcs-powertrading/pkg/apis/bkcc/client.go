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

// Package bkcc xxx
package bkcc

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis/requester"
)

// Client interface
type Client interface {
	ListHostByCC(ctx context.Context, ipList []string, bizId string) ([]CCHostInfo, error)
}

const (
	listBizHosts = "%s/api/c/compapi/v2/cc/list_biz_hosts/"
)

// New client
func New(opt *apis.ClientOptions, r requester.Requester) Client {
	cli := &client{
		requestClient: r,
		opt:           opt,
	}
	err := cli.SetDefaultHeader()
	if err != nil {
		return nil
	}
	return cli
}

type client struct {
	defaultHeader map[string]string
	requestClient requester.Requester
	opt           *apis.ClientOptions
}

// SetDefaultHeader set default header
func (c *client) SetDefaultHeader() error {
	header := make(map[string]string)
	auth := &apis.BkAuthOpts{
		BkAppCode:   c.opt.AppCode,
		BkAppSecret: c.opt.AppSecret,
		AccessToken: c.opt.AccessToken,
	}
	str, err := json.Marshal(auth)
	if err != nil {
		blog.Errorf("marshal header error:%s", err.Error())
		return err
	}
	header["X-Bkapi-Authorization"] = string(str)
	c.defaultHeader = header
	return nil
}

// ListHostByCC list host by cc
func (c *client) ListHostByCC(ctx context.Context, ipList []string, bizId string) ([]CCHostInfo, error) {
	if len(ipList) == 0 {
		return nil, errors.Errorf("cc list hosts ips cannot be empty")
	}
	biz, err := strconv.Atoi(bizId)
	if err != nil {
		return nil, errors.Errorf("trans bizId string to int64 failed:%s", err.Error())
	}
	request := &listBizHostsRequest{
		appInfo: appInfo{
			AppCode:   c.opt.AppCode,
			AppSecret: c.opt.AppSecret,
			Operator:  "bcs",
		},
		Page: &page{
			Start: 0,
			Limit: len(ipList),
		},
		BkBizID: int64(biz),
		Fields: []string{"bk_host_id", "bk_cloud_id", "bk_host_innerip", "bk_os_type", "bk_mac",
			"bk_asset_id", "idc_name", "idc_unit_name", "idc_city_name"},
		HostPropertyFilter: &hostPropertyFilter{
			Condition: "AND",
			Rules: []*queryRule{
				{
					Field:    "bk_host_innerip",
					Operator: "in",
					Value:    ipList,
				},
			},
		},
	}
	reqData, err := json.Marshal(request)
	if err != nil {
		blog.Errorf("Error encoding JSON: %v", err)
	}
	url := fmt.Sprintf(listBizHosts, c.opt.Endpoint)
	rsp, requestErr := c.requestClient.DoPostRequest(url, c.defaultHeader, reqData)
	if requestErr != nil {
		return nil, fmt.Errorf("do listHost request error:%s", requestErr.Error())
	}
	result := &listHostsWithoutBizResponse{}
	unMarshalErr := json.Unmarshal(rsp, &result)
	if unMarshalErr != nil {
		return nil, fmt.Errorf("do unmarshal error, url: %s, error:%v", url, unMarshalErr.Error())
	}

	if result.Code != 0 {
		return nil, errors.Errorf("response code not 0 but %d: %s", result.Code, result.Message)
	}
	if result.Data != nil {
		return result.Data.Info, nil
	}
	return nil, nil
}

// ListHostWithoutBiz list host without biz
func (c *client) ListHostWithoutBiz(ctx context.Context, ipList []string) ([]CCHostInfo, error) {
	if len(ipList) == 0 {
		return nil, errors.Errorf("cc list hosts ips cannot be empty")
	}
	request := &listBizHostsRequest{
		appInfo: appInfo{
			AppCode:   c.opt.AppCode,
			AppSecret: c.opt.AppSecret,
			Operator:  "bcs",
		},
		Page: &page{
			Start: 0,
			Limit: len(ipList),
		},
		Fields: []string{"bk_host_id", "bk_cloud_id", "bk_host_innerip", "bk_os_type", "bk_mac",
			"bk_asset_id", "idc_name", "idc_unit_name", "idc_city_name"},
		HostPropertyFilter: &hostPropertyFilter{
			Condition: "AND",
			Rules: []*queryRule{
				{
					Field:    "bk_host_innerip",
					Operator: "in",
					Value:    ipList,
				},
			},
		},
	}
	reqData, err := json.Marshal(request)
	if err != nil {
		blog.Errorf("Error encoding JSON: %v", err)
	}
	url := fmt.Sprintf(listBizHosts, c.opt.Endpoint)
	rsp, requestErr := c.requestClient.DoPostRequest(url, c.defaultHeader, reqData)
	if requestErr != nil {
		return nil, fmt.Errorf("do listHost request error:%s", requestErr.Error())
	}
	result := &listHostsWithoutBizResponse{}
	unMarshalErr := json.Unmarshal(rsp, &result)
	if unMarshalErr != nil {
		return nil, fmt.Errorf("do unmarshal error, url: %s, error:%v", url, unMarshalErr.Error())
	}

	if result.Code != 0 {
		return nil, errors.Errorf("response code not 0 but %d: %s", result.Code, result.Message)
	}
	if result.Data != nil {
		return result.Data.Info, nil
	}
	return nil, nil
}
