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
	"crypto/tls"
	"net/http"

	paasclient "github.com/Tencent/bk-bcs/bcs-common/pkg/esb/client"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/throttle"
)

// ClientInterface client interface for cmdb
type ClientInterface interface {
	// container server
	CreatePod(bizID int64, data *CreatePod) (*CreatedOneOptionResult, error)
	CreateManyPod(bizID int64, data *CreateManyPod) (*CreatedManyOptionResult, error)
	UpdatePod(bizID int64, data *UpdatePod) (*UpdatedOptionResult, error)
	DeletePod(bizID int64, data *DeletePod) (*DeletedOptionResult, error)
	ListClusterPods(bizID int64, clusterID string) (*ListPodsResult, error)
	// topo server
	SearchBusinessTopoWithStatistics(bizID int64) (*SearchBusinessTopoWithStatisticsResult, error)

	// interfaces through esb
	// 如果是通过ESB的请求，所有的参数都在post请求体中
	ESBSearchBusiness(username string, condition map[string]interface{}) (*ESBSearchBusinessResult, error)
	ESBSearchBizInstTopo(username string, bizID int64) (*ESBSearchBizInstTopoResult, error)
	ESBTransferHostInBizModule(username string, bizID int64, hostIDs, moduleIDs []int64) (
		*ESBTransferHostModuleResult, error)
	ESBListHostsWithoutBiz(username string, req *ESBListHostsWitoutBizRequest) (*ESBListHostsWitoutBizResult, error)
	ESBGetBizLocation(username string, bizIDs []int64) (*ESBGetBizLocationResult, error)
	ESBGetBizInternalModule(usename string, bizID int64, bkSupplierAccount string) (*ESBGetBizInternalModuleResult, error)
	ESBListBizHosts(username string, req *ESBListBizHostsRequest) (*ESBListBizHostsResult, error)
	ESBListBizHostsTopo(username string, req *ESBListBizHostsTopoRequest) (*ESBListBizHostsTopoResult, error)
}

// NewClientInterface create client interface
func NewClientInterface(host string, tlsConf *tls.Config) *Client {
	var cli *paasclient.RESTClient
	if tlsConf != nil {
		cli = paasclient.NewRESTClientWithTLS(tlsConf).
			WithRateLimiter(throttle.NewTokenBucket(1000, 1000))
	} else {
		cli = paasclient.NewRESTClient().
			WithRateLimiter(throttle.NewTokenBucket(1000, 1000))
	}

	return &Client{
		host:    host,
		client:  cli,
		baseReq: make(map[string]interface{}),
	}
}

// Client paas cmdb client
type Client struct {
	host          string
	defaultHeader http.Header
	client        *paasclient.RESTClient
	baseReq       map[string]interface{}
}

// SetDefaultHeader set default headers
func (c *Client) SetDefaultHeader(h http.Header) {
	c.defaultHeader = h
}

// SetCommonReq set base req
func (c *Client) SetCommonReq(args map[string]interface{}) {
	for k, v := range args {
		c.baseReq[k] = v
	}
}

// CreatePod create pod
func (c *Client) CreatePod(bizID int64, data *CreatePod) (*CreatedOneOptionResult, error) {
	result := new(CreatedOneOptionResult)
	req := map[string]interface{}{
		"pod": data.Pod,
	}
	err := c.client.Post().
		WithEndpoints([]string{c.host}).
		WithBasePath("/api/v3/").
		SubPathf("/create/container/bk_biz_id/%d/pod", bizID).
		WithHeaders(c.defaultHeader).
		Body(req).
		Do().
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// CreateManyPod create many pod
func (c *Client) CreateManyPod(bizID int64, data *CreateManyPod) (*CreatedManyOptionResult, error) {
	result := new(CreatedManyOptionResult)
	req := map[string]interface{}{
		"pod": data.PodList,
	}
	err := c.client.Post().
		WithEndpoints([]string{c.host}).
		WithBasePath("/api/v3/").
		SubPathf("createmany/container/bk_biz_id/%d/pod", bizID).
		WithHeaders(c.defaultHeader).
		Body(req).
		Do().
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// UpdatePod update pod
func (c *Client) UpdatePod(bizID int64, data *UpdatePod) (*UpdatedOptionResult, error) {
	result := new(UpdatedOptionResult)
	req := map[string]interface{}{
		"condition": data.Condition,
		"data":      data.Data,
	}
	err := c.client.Put().
		WithEndpoints([]string{c.host}).
		WithBasePath("/api/v3/").
		SubPathf("update/container/bk_biz_id/%d/pod", bizID).
		WithHeaders(c.defaultHeader).
		Body(req).
		Do().
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// DeletePod delete pod
func (c *Client) DeletePod(bizID int64, data *DeletePod) (*DeletedOptionResult, error) {
	result := new(DeletedOptionResult)
	req := map[string]interface{}{
		"condition": data.Condition,
	}
	err := c.client.Delete().
		WithEndpoints([]string{c.host}).
		WithBasePath("/api/v3/").
		SubPathf("delete/container/bk_biz_id/%d/pod", bizID).
		WithHeaders(c.defaultHeader).
		Body(req).
		Do().
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ListClusterPods list cluster pods
func (c *Client) ListClusterPods(bizID int64, clusterID string) (*ListPodsResult, error) {
	request := map[string]interface{}{
		"bk_biz_id": bizID,
		"pod_property_filter": map[string]interface{}{
			"condition": "AND",
			"rules": []map[string]interface{}{
				{
					"field":    "bk_pod_cluster",
					"operator": "equal",
					"value":    clusterID,
				},
			},
		},
	}
	result := new(ListPodsResult)
	err := c.client.Post().
		WithEndpoints([]string{c.host}).
		WithBasePath("/api/v3/").
		SubPathf("findmany/container/bk_biz_id/%d/pod", bizID).
		WithHeaders(c.defaultHeader).
		Body(request).
		Do().
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// SearchBusinessTopoWithStatistics implements client interface
func (c *Client) SearchBusinessTopoWithStatistics(bizID int64) (*SearchBusinessTopoWithStatisticsResult, error) {
	result := new(SearchBusinessTopoWithStatisticsResult)
	err := c.client.Post().
		WithEndpoints([]string{c.host}).
		WithBasePath("/api/v3/").
		SubPathf("find/topoinst_with_statistics/biz/%d", bizID).
		WithHeaders(c.defaultHeader).
		Do().
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
