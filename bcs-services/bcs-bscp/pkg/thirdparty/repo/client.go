/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package repo

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"bscp.io/pkg/cc"
	"bscp.io/pkg/rest"
	"bscp.io/pkg/rest/client"
	"bscp.io/pkg/tools"

	"github.com/prometheus/client_golang/prometheus"
)

// Client is repo client.
type Client struct {
	config *cc.Repository
	// http client instance
	client rest.ClientInterface
	// http header info
	basicHeader http.Header
}

// NewClient new repo client.
func NewClient(repoSetting *cc.Repository, reg prometheus.Registerer) (*Client, error) {
	tls := &tools.TLSConfig{
		InsecureSkipVerify: repoSetting.TLS.InsecureSkipVerify,
		CertFile:           repoSetting.TLS.CertFile,
		KeyFile:            repoSetting.TLS.KeyFile,
		CAFile:             repoSetting.TLS.CAFile,
		Password:           repoSetting.TLS.Password,
	}
	cli, err := client.NewClient(tls)
	if err != nil {
		return nil, err
	}

	c := &client.Capability{
		Client: cli,
		Discover: &repoDiscovery{
			servers: repoSetting.Endpoints,
		},
		MetricOpts: client.MetricOption{Register: reg},
	}

	header := http.Header{}
	header.Set("Content-Type", "application/json")
	header.Set("Accept", "application/json")
	header.Set("Authorization", fmt.Sprintf("Platform %s", repoSetting.Token))
	header.Set(HeaderKeyUID, repoSetting.User)

	return &Client{
		config:      repoSetting,
		client:      rest.NewClient(c, "/"),
		basicHeader: header,
	}, nil
}

// ProjectID return repo project id.
func (c *Client) ProjectID() string {
	return c.config.Project
}

// IsProjectExist judge repo bscp project already exist.
func (c *Client) IsProjectExist(ctx context.Context) error {
	resp := c.client.Get().
		WithContext(ctx).
		SubResourcef("/repository/api/project/exist/%s", c.config.Project).
		WithHeaders(c.basicHeader).
		Do()
	if resp.Err != nil {
		return resp.Err
	}

	// repo uses StatusBadRequest to mark the failure of the request, so StatusBadRequest
	// needs special handling to read out the error information in the body message.
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest {
		return fmt.Errorf("response status code: %d", resp.StatusCode)
	}

	respBody := new(rest.BaseResp)
	if err := resp.Into(respBody); err != nil {
		return err
	}

	if respBody.Code != 0 {
		return fmt.Errorf("code: %d, message: %s", respBody.Code, respBody.Message)
	}

	return nil
}

// CreateRepo create new repository in repo.
func (c *Client) CreateRepo(ctx context.Context, req *CreateRepoReq) error {
	resp := c.client.Post().
		WithContext(ctx).
		SubResourcef("/repository/api/repo/create").
		WithHeaders(c.basicHeader).
		Body(req).
		Do()
	if resp.Err != nil {
		return resp.Err
	}

	// repo uses StatusBadRequest to mark the failure of the request, so StatusBadRequest
	// needs special handling to read out the error information in the body message.
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest {
		return fmt.Errorf("response status code: %d", resp.StatusCode)
	}

	respBody := new(rest.BaseResp)
	if err := resp.Into(respBody); err != nil {
		return err
	}

	// if repo already exist, ignore this error.
	if respBody.Code != 0 && respBody.Code != errCodeRepoAlreadyExist {
		return fmt.Errorf("code: %d, message: %s", respBody.Code, respBody.Message)
	}

	return nil
}

// DeleteRepo delete repository in repo. param force: whether to force deletion.
// If false, the warehouse cannot be deleted when there are files in the warehouse
func (c *Client) DeleteRepo(ctx context.Context, bizID uint32, forced bool) error {
	repoName, err := GenRepoName(bizID)
	if err != nil {
		return err
	}

	resp := c.client.Delete().
		WithContext(ctx).
		SubResourcef("/repository/api/repo/delete/%s/%s", c.config.Project, repoName).
		WithParam("forced", strconv.FormatBool(forced)).
		WithHeaders(c.basicHeader).
		Do()
	if resp.Err != nil {
		return resp.Err
	}

	// repo uses StatusBadRequest to mark the failure of the request, so StatusBadRequest
	// needs special handling to read out the error information in the body message.
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest {
		return fmt.Errorf("response status code: %d", resp.StatusCode)
	}

	respBody := new(rest.BaseResp)
	if err := resp.Into(respBody); err != nil {
		return err
	}

	if respBody.Code != 0 {
		return fmt.Errorf("code: %d, message: %s", respBody.Code, respBody.Message)
	}

	return nil
}

// IsNodeExist judge repo node already exist.
func (c *Client) IsNodeExist(ctx context.Context, nodePath string) (bool, error) {
	resp := c.client.Head().
		WithContext(ctx).
		SubResourcef(nodePath).
		WithHeaders(c.basicHeader).
		Do()
	if resp.Err != nil {
		return false, resp.Err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		return false, fmt.Errorf("response status code: %d", resp.StatusCode)
	}

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	return true, nil
}

// DeleteNode delete node.
func (c *Client) DeleteNode(ctx context.Context, nodePath string) error {
	resp := c.client.Delete().
		WithContext(ctx).
		SubResourcef(nodePath).
		WithHeaders(c.basicHeader).
		Do()
	if resp.Err != nil {
		return resp.Err
	}

	// repo uses StatusBadRequest to mark the failure of the request, so StatusBadRequest
	// needs special handling to read out the error information in the body message.
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest {
		return fmt.Errorf("response status code: %d", resp.StatusCode)
	}

	respBody := new(rest.BaseResp)
	if err := resp.Into(respBody); err != nil {
		return err
	}

	if respBody.Code != 0 {
		return fmt.Errorf("code: %d, message: %s", respBody.Code, respBody.Message)
	}
	return nil
}

// QueryMetaResp query metadata repo return response.
type QueryMetaResp struct {
	Code    uint32            `json:"code"`
	Message string            `json:"message"`
	Data    map[string]string `json:"data"`
}

// QueryMetadata query node metadata info. If node not exist, return data is {}.
func (c *Client) QueryMetadata(ctx context.Context, opt *NodeOption) (map[string]string, error) {
	repoName, err := GenRepoName(opt.BizID)
	if err != nil {
		return nil, err
	}

	fullPath, err := GenNodeFullPath(opt.Sign)
	if err != nil {
		return nil, err
	}

	resp := c.client.Get().
		WithContext(ctx).
		SubResourcef("/repository/api/metadata/%s/%s%s", opt.Project, repoName, fullPath).
		WithHeaders(c.basicHeader).
		Do()
	if resp.Err != nil {
		return nil, resp.Err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response status code: %d", resp.StatusCode)
	}

	respBody := new(QueryMetaResp)
	if err := resp.Into(respBody); err != nil {
		return nil, err
	}

	if respBody.Code != 0 {
		return nil, fmt.Errorf("code: %d, message: %s", respBody.Code, respBody.Message)
	}

	return respBody.Data, nil
}
