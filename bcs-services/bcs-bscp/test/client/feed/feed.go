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

// Package feed NOTES
package feed

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/cmd/feed-server/bll/types"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/rest"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/rest/client"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

// Client feed server client
type Client struct {
	feedClient rest.ClientInterface
}

// NewFeedClient get a new feed client
func NewFeedClient(host string, c *tools.TLSConfig) (*Client, error) {
	httpCli, err := client.NewClient(c)
	if err != nil {
		return nil, err
	}

	capCli := &client.Capability{
		Client: httpCli,
		Discover: &discovery{
			server: host,
		},
		ToleranceLatencyTime: 2 * time.Second,
		MetricOpts:           client.MetricOption{Register: nil},
	}

	return &Client{
		feedClient: rest.NewClient(capCli, "/api/v1/feed"),
	}, nil
}

// ListAppFileLatestRelease list application file's latest release
func (c *Client) ListAppFileLatestRelease(ctx context.Context, header http.Header,
	req *types.ListFileAppLatestReleaseMetaReq) (*types.ListFileAppLatestReleaseMetaResp, error) {

	resp := c.feedClient.Post().
		WithContext(ctx).
		SubResourcef("/list/app/release/type/file/latest").
		WithHeaders(header).
		Body(req).
		Do()

	pbResp := new(types.ListFileAppLatestReleaseMetaResp)
	if err := resp.Into(pbResp); err != nil {
		return nil, err
	}

	return pbResp, nil
}

type discovery struct {
	server string
}

// GetServers get feed severs
func (r *discovery) GetServers() ([]string, error) {
	if len(r.server) == 0 {
		return nil, errors.New("can not get feed server")
	}
	return []string{r.server}, nil
}
