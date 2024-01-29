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

package api

import (
	"context"
	"net/http"
	"reflect"

	pbcs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/config-server"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/rest"
)

// Commit commit client
type Commit struct {
	client rest.ClientInterface
}

// NewCommitClient get a new commit client
func NewCommitClient(client rest.ClientInterface) *Commit {
	return &Commit{
		client: client,
	}
}

// Create function to create commit.
func (c *Commit) Create(ctx context.Context, header http.Header, req *pbcs.CreateCommitReq) (
	*pbcs.CreateCommitResp, error) {
	resp := c.client.Post().
		WithContext(ctx).
		SubResourcef("/config/create/commit/commit/config_item_id/%d/app_id/%d/biz_id/%d",
			req.ConfigItemId, req.AppId, req.BizId).
		WithHeaders(header).
		Body(req).
		Do()

	if resp.Err != nil {
		return nil, resp.Err
	}

	pbResp := &struct {
		Data  *pbcs.CreateCommitResp `json:"data"`
		Error *rest.ErrorPayload     `json:"error"`
	}{}
	if err := resp.Into(pbResp); err != nil {
		return nil, err
	}
	if !reflect.ValueOf(pbResp.Error).IsNil() {
		return nil, pbResp.Error
	}

	return pbResp.Data, nil
}

// List to list commit.
func (c *Commit) List(ctx context.Context, header http.Header, req *pbcs.ListCommitsReq) (
	*pbcs.ListCommitsResp, error) {
	resp := c.client.Post().
		WithContext(ctx).
		SubResourcef("/config/list/commit/commit/app_id/%d/biz_id/%d", req.AppId, req.BizId).
		WithHeaders(header).
		Body(req).
		Do()

	if resp.Err != nil {
		return nil, resp.Err
	}

	pbResp := &struct {
		Data  *pbcs.ListCommitsResp `json:"data"`
		Error *rest.ErrorPayload    `json:"error"`
	}{}
	if err := resp.Into(pbResp); err != nil {
		return nil, err
	}
	if !reflect.ValueOf(pbResp.Error).IsNil() {
		return nil, pbResp.Error
	}

	return pbResp.Data, nil
}
