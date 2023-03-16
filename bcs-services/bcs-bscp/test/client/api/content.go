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

package api

import (
	"context"
	"net/http"

	pbcs "bscp.io/pkg/protocol/config-server"
	"bscp.io/pkg/rest"
)

// Content content client
type Content struct {
	client rest.ClientInterface
}

// NewContentClient get a new content client
func NewContentClient(client rest.ClientInterface) *Content {
	return &Content{
		client: client,
	}
}

// Create function to create content.
func (c *Content) Create(ctx context.Context, header http.Header, req *pbcs.CreateContentReq) (
	*pbcs.CreateContentResp, error) {

	resp := c.client.Post().
		WithContext(ctx).
		SubResourcef("/config/create/content/content/config_item_id/%d/app_id/%d/biz_id/%d",
			req.ConfigItemId, req.AppId, req.BizId).
		WithHeaders(header).
		Body(req).
		Do()

	if resp.Err != nil {
		return nil, resp.Err
	}

	pbResp := &struct {
		Data *pbcs.CreateContentResp `json:"data"`
	}{}
	if err := resp.Into(pbResp); err != nil {
		return nil, err
	}

	return pbResp.Data, nil
}

// Upload function to upload content.
func (c *Content) Upload(ctx context.Context, header http.Header, bizId, appId uint32, data string) (
	*rest.BaseResp, error) {

	resp := c.client.Put().
		WithContext(ctx).
		SubResourcef("/api/create/content/upload/biz_id/%d/app_id/%d", bizId, appId).
		WithHeaders(header).
		Body(data).
		Do()

	if resp.Err != nil {
		return nil, resp.Err
	}

	pbResp := new(rest.BaseResp)
	if err := resp.Into(pbResp); err != nil {
		return nil, err
	}

	return pbResp, nil
}

// List to list contents.
func (c *Content) List(ctx context.Context, header http.Header, req *pbcs.ListContentsReq) (
	*pbcs.ListContentsResp, error) {

	resp := c.client.Post().
		WithContext(ctx).
		SubResourcef("/config/list/content/content/app_id/%d/biz_id/%d", req.AppId, req.BizId).
		WithHeaders(header).
		Body(req).
		Do()

	if resp.Err != nil {
		return nil, resp.Err
	}

	pbResp := &struct {
		Data *pbcs.ListContentsResp `json:"data"`
	}{}
	if err := resp.Into(pbResp); err != nil {
		return nil, err
	}

	return pbResp.Data, nil
}
