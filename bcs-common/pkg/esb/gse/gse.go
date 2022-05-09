/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package gse

import (
	"crypto/tls"
	"net/http"

	paasclient "github.com/Tencent/bk-bcs/bcs-common/pkg/esb/client"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/throttle"
)

// ClientInterface client interface for gse
type ClientInterface interface {
	GetAgentStatus(req *GetAgentStatusReq) (*GetAgentStatusResp, error)
}

// NewClient create client
func NewClient(host string, tlsConf *tls.Config) *Client {
	var cli *paasclient.RESTClient
	if tlsConf != nil {
		cli = paasclient.NewRESTClientWithTLS(tlsConf).
			WithRateLimiter(throttle.NewTokenBucket(1000, 1000))
	} else {
		cli = paasclient.NewRESTClient().
			WithRateLimiter(throttle.NewTokenBucket(1000, 1000))
	}

	return &Client{
		host:   host,
		client: cli,
	}
}

// Client client for gse service
type Client struct {
	host          string
	defaultHeader http.Header
	client        *paasclient.RESTClient
	baseReq       BaseReq
}

// SetDefaultHeader set default headers
func (c *Client) SetDefaultHeader(h http.Header) {
	c.defaultHeader = h
}

// SetCommonReq set base request
func (c *Client) SetCommonReq(br BaseReq) {
	c.baseReq = br
}

// GetAgentStatus get gse agent status
func (c *Client) GetAgentStatus(req *GetAgentStatusReq) (*GetAgentStatusResp, error) {
	resp := new(GetAgentStatusResp)
	req.BaseReq = c.baseReq
	err := c.client.Post().
		WithEndpoints([]string{c.host}).
		WithBasePath("/api/c/compapi/v2/gse/").
		SubPathf("get_agent_status").
		WithJSON(req).
		Do().
		Into(resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
