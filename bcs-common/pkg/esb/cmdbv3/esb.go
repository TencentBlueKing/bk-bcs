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

package cmdbv3

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/esb/common"
)

// ESBSearchBusiness search business
func (c *Client) ESBSearchBusiness(username string, condition map[string]interface{}) (
	*ESBSearchBusinessResult, error) {

	request := map[string]interface{}{
		"condition":   condition,
		"bk_username": username,
	}
	common.MergeMap(request, c.baseReq)
	result := new(ESBSearchBusinessResult)
	err := c.client.Post().
		WithEndpoints([]string{c.host}).
		WithBasePath("/api/c/compapi/v2/cc/").
		SubPathf("search_business").
		WithHeaders(c.defaultHeader).
		Body(request).
		Do().
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ESBTransferHostInBizModule transfer host in biz
func (c *Client) ESBTransferHostInBizModule(username string, bizID int64, hostIDs, moduleIDs []int64) (
	*ESBTransferHostModuleResult, error) {

	request := map[string]interface{}{
		"bk_biz_id":    bizID,
		"bk_username":  username,
		"bk_host_id":   hostIDs,
		"bk_module_id": moduleIDs,
		"is_increment": false,
	}
	common.MergeMap(request, c.baseReq)
	result := new(ESBTransferHostModuleResult)
	err := c.client.Post().
		WithEndpoints([]string{c.host}).
		WithBasePath("/api/c/compapi/v2/cc/").
		SubPathf("transfer_host_module").
		WithHeaders(c.defaultHeader).
		Body(request).
		Do().
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ESBSearchBizInstTopo search biz instance topo
func (c *Client) ESBSearchBizInstTopo(username string, bizID int64) (*ESBSearchBizInstTopoResult, error) {
	request := map[string]interface{}{
		"bk_biz_id":   bizID,
		"bk_username": username,
	}
	common.MergeMap(request, c.baseReq)
	result := new(ESBSearchBizInstTopoResult)
	err := c.client.Post().
		WithEndpoints([]string{c.host}).
		WithBasePath("/api/c/compapi/v2/cc/").
		SubPathf("search_biz_inst_topo").
		WithHeaders(c.defaultHeader).
		Body(request).
		Do().
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ESBListHostsWithoutBiz list hosts without biz
func (c *Client) ESBListHostsWithoutBiz(username string, req *ESBListHostsWitoutBizRequest) (
	*ESBListHostsWitoutBizResult, error) {

	if req == nil {
		return nil, fmt.Errorf("request is empty")
	}
	request := map[string]interface{}{
		"bk_username": username,
	}
	if req.BkBizID != 0 {
		request["bk_biz_id"] = req.BkBizID
	}
	if req.HostPropertyFilter != nil {
		request["host_property_filter"] = req.HostPropertyFilter
	}
	if req.Page != nil {
		request["page"] = req.Page
	}
	result := new(ESBListHostsWitoutBizResult)
	common.MergeMap(request, c.baseReq)
	err := c.client.Post().
		WithEndpoints([]string{c.host}).
		WithBasePath("/api/c/compapi/v2/cc/").
		SubPathf("list_hosts_without_biz").
		WithHeaders(c.defaultHeader).
		Body(request).
		Do().
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ESBGetBizLocation get biz location
func (c *Client) ESBGetBizLocation(username string, bizIDs []int64) (*ESBGetBizLocationResult, error) {
	if len(bizIDs) == 0 {
		return nil, fmt.Errorf("bk_biz_ids cannot be empty")
	}
	request := map[string]interface{}{
		"bk_username": username,
		"bk_biz_ids":  bizIDs,
	}
	common.MergeMap(request, c.baseReq)
	result := new(ESBGetBizLocationResult)
	err := c.client.Post().
		WithEndpoints([]string{c.host}).
		WithBasePath("/api/c/compapi/v2/cc/").
		SubPathf("get_biz_location").
		WithHeaders(c.defaultHeader).
		Body(request).
		Do().
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
