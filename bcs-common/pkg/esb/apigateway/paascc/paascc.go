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

// Package paascc xxx
package paascc

import (
	"crypto/tls"

	paasclient "github.com/Tencent/bk-bcs/bcs-common/pkg/esb/client"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/throttle"
)

// ClientInterface client interface for paas-cc
type ClientInterface interface {
	ListProjects(env string) (*ListProjectsResult, error)
	ListProjectClusters(env, projectID string) (*ListProjectClustersResult, error)
	GetProject(env, projectid string) (*GetProjectResult, error)
}

// NewClientInterface create client interface
func NewClientInterface(host, appcode, appsecret string, tlsConf *tls.Config) ClientInterface {
	var cli *paasclient.RESTClient
	if tlsConf != nil {
		cli = paasclient.NewRESTClientWithTLS(tlsConf).
			WithRateLimiter(throttle.NewTokenBucket(50, 50)).
			WithCredential(appcode, appsecret)
	} else {
		cli = paasclient.NewRESTClient().
			WithRateLimiter(throttle.NewTokenBucket(50, 50)).
			WithCredential(appcode, appsecret)
	}

	return &Client{
		host:   host,
		client: cli,
	}
}

// Client paas cc client
type Client struct {
	host   string
	client *paasclient.RESTClient
	// reserved for BK PaaS AppToken, setting formation in http header:
	// X-Bkapi-Authorization: '{"access_token": "xxxxxxxxxxxxxxxxxxxx"}'
	accessToken string // nolint
}

// ListProjects list projects
func (c *Client) ListProjects(env string) (*ListProjectsResult, error) {
	result := &ListProjectsResult{}
	err := c.client.Get().
		WithEndpoints([]string{c.host}).
		WithBasePath("/").
		SubPathf("%s/projects/", env).
		WithParam("desire_all_data", "1").
		Do().
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetProject get project
func (c *Client) GetProject(env, projectid string) (*GetProjectResult, error) {
	result := &GetProjectResult{}
	err := c.client.Get().
		WithEndpoints([]string{c.host}).
		WithBasePath("/").
		SubPathf("%s/projects/%s", env, projectid).
		Do().
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ListProjectClusters list project clusters
func (c *Client) ListProjectClusters(env, projectID string) (*ListProjectClustersResult, error) {
	result := &ListProjectClustersResult{}
	err := c.client.Get().
		WithEndpoints([]string{c.host}).
		WithBasePath("/").
		SubPathf("/%s/projects/%s/clusters", env, projectID).
		Do().
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
